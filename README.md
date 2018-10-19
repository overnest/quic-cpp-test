This repo has been updated to use a Singleton design pattern that has two purposes:
1. To prevent duplicate GoLang containers.
2. To allow external calls to the library without dealing with Go types.

Since Go cannot export complex structs such as network connections and socket listeners to C++, I was forced to keep them inside the Go container like this:
```Go
var ln net.Listener
var conns sync.Map
```
By keeping these global they do not need to be included in function parameters and return types which allow them to exported with cgo. Since C++ can't see any of this happening calling the exported library must be done from a Singleton.

Look inside the references folder to find the GoLang file that was used to generate libquic.so and libquic.h.
The folder also includes instructions to generate the files yourself.

Example instructions:
1. `$make server` and `$make client` or just run `$make all`
2. Run the binaries with `./server` and `./client`

To include in another project you will need 4 files:
1. QUIC.cpp
2. QUIC.h
3. libquic.so -> /usr/lib
4. libquic.h

The example client and server use a bare bones version of stellar-core's PeerDoor class and another QUICPeer class that I created. I integrated QUIC into stellar-core by doing the following:
1. Copy QUIC.cpp, QUIC.h, and libquic.h into src/overlay
2. Copy libquic.so into /usr/lib
3. Create QUICPeer.cpp and QUICPeer.h to inherit from Peer.h the same way as TCPPeer does. Modify member functions as necesary.
4. Modify PeerDoor.h to hold the QUIC* and PeerDoor to listen for incoming connections. Example below.
5. Modify OverlayManagerImpl.cpp to try to connect using QUIC first, then use TCP if that fails.
6. To tell make to include new files just include them to git: `git add QUICPeer.h QUICPeer.cpp QUIC.h QUIC.cpp libquic.h`. Then just re-run `./autogen.sh` and `./configure` to generate an updated Makefile.
7. Add the -lquic flag into the CXXFLAGS in the new src/Makefile so it will use the shared library. Example below

src/Makefile
```Makefile
CXXFLAGS = -g -O2  -lquic -pthread -Wall ...
```
src/overlay/PeerDoor.cpp
```C++
...

void
PeerDoor::start()
{
    if (!mApp.getConfig().RUN_STANDALONE)
    {
        quic_ptr = QUIC::getInstance();
        auto callback = std::bind(&PeerDoor::quicKnock, this,placeholders::_1);
        quic_ptr->start(mApp.getConfig().PEER_PORT+1, callback);
	...
    }
}

...

std::function<void(const char*)>
PeerDoor::quicKnock(int ID)
{
    CLOG(DEBUG, "Overlay") << "PeerDoor quicKnock() @"
                           << mApp.getConfig().PEER_PORT+1;
    QUICPeer::pointer peer = QUICPeer::accept(mApp, ID);
    if (peer)
    {
        mApp.getOverlayManager().addPendingPeer(peer);
    }
    auto callback = bind(&QUICPeer::recvMessage,peer.get(), placeholders::_1);
    return callback;
}
```
PeerDoor::start needs to initiate listening for both TCP and QUIC connections. You will also need to create PeerDoor::quicKnoc(int ID) to handle incoming connections. In this case I used std::bind and std::functional to return the receive function of the new Peer. You will also need to make a few small changes to PeerDoor.h:

src/overlay/PeerDoor.h
```C++
...

#include <functional>

#include "overlay/QUIC.h"

...

private:
    QUIC* quic_ptr;
    
public:
    std::function<void(const char*)> quicKnock(int ID);
    
...

```
src/overlay/OverlayManagerImpl.cpp
```C++
...

#include "QUICPeer.h"

...

void
OverlayManagerImpl::connectTo(PeerRecord& pr)
{
    mConnectionsAttempted.Mark();
    if (!getConnectedPeer(pr.getAddress()))
    {
        pr.backOff(mApp.getClock());
        pr.storePeerRecord(mApp.getDatabase());

        if (getPendingPeersCount() < mApp.getConfig().MAX_PENDING_CONNECTIONS)
        {
            auto quic_peer = QUICPeer::initiate(mApp, pr.getAddress());
            if(quic_peer->quic_id < 0){
                addPendingPeer(TCPPeer::initiate(mApp, pr.getAddress()));
            }else{
                addPendingPeer(quic_peer);
            }
        }
        ...
    }
    ...
}
```
OverlayManagerImpl.cpp is where connections are initiated. quic_id will be -1 if unable to connect, so thats how we know to use TCP instead. No need to update the header file. Just make sure to include the new QUICPeer.h


The following two files are a WIP. The goal is for TCPPeer and QUICPeer to be both valid implementations of Peer.
src/overlay/QUICPeer.cpp
```C++
QUICPeer::pointer
QUICPeer::initiate(Application& app, PeerBareAddress const& address)
{
    assert(address.getType() == PeerBareAddress::Type::IPv4);

    CLOG(DEBUG, "Overlay") << "QUICPeer:initiate"
                           << " to " << address.toString();
    auto result = make_shared<QUICPeer>(app, WE_CALLED_REMOTE);
    result->mAddress = address;
    result->quic_ptr = QUIC::getInstance();
    auto callback =  bind(&QUICPeer::recvMessage,result.get(),placeholders::_1);
    result->quic_id = result->quic_ptr->connect(address.getIP().c_str(), address.getPort()+1, callback);
    return result;
}

QUICPeer::pointer
QUICPeer::accept(Application& app, int ID)
{
    CLOG(DEBUG, "Overlay") << "QUICPeer:accept"
                           << "@" << app.getConfig().PEER_PORT;
    auto result = make_shared<QUICPeer>(app, REMOTE_CALLED_US);
    result->startIdleTimer();
    
    result->quic_ptr = QUIC::getInstance();
    result->quic_id = ID;

    return result;
}

QUICPeer::~QUICPeer()
{
    quic_ptr->disconnect(quic_id);
    mIdleTimer.cancel();
}

PeerBareAddress
QUICPeer::makeAddress(int remoteListeningPort) const
{
    if (quic_id < 0)
    {
        return PeerBareAddress{};
    }
    else
    {
        return PeerBareAddress{
            quic_ptr->getAddr(quic_id),
            static_cast<unsigned short>(remoteListeningPort)};
    }
}

void
QUICPeer::sendMessage(xdr::msg_ptr&& xdrBytes)
{
    if (Logging::logTrace("Overlay"))
        CLOG(TRACE, "Overlay") << "QUICPeer:sendMessage to " << toString();

    // places the buffer to write into the write queue
    auto buf = std::make_shared<xdr::msg_ptr>(std::move(xdrBytes));

    if(quic_id >= 0){
        bool success = quic_ptr->sendMsg(quic_id, (*buf)->raw_data());
        if (success)
            return;
    }
}

void
QUICPeer::recvMessage(const char *str)
{
    if(str == NULL)
    {
        drop(true);
        return;
    }
    try
    {
        int len = 0;
        while(str[len] != '\0') len++;
        xdr::xdr_get g(str, str + len * sizeof(char));
        AuthenticatedMessage am;
        xdr::xdr_argpack_archive(g, am);
        Peer::recvMessage(am);
    }
    catch (xdr::xdr_runtime_error& e)
    {
        CLOG(ERROR, "Overlay") << "recvMessage got a corrupt xdr: " << e.what();
    }
}

void
QUICPeer::drop(bool force)
{
    if (shouldAbort())
    {
        return;
    }

    CLOG(DEBUG, "Overlay") << "TCPPeer::drop " << toString() << " in state "
                           << mState << " we called:" << mRole;

    mState = CLOSING;

    auto self = static_pointer_cast<QUICPeer>(shared_from_this());
    getApp().getOverlayManager().dropPeer(this);

    // if write queue is not empty, messageSender will take care of shutdown
    if (force || !mWriting)
    {
        quic_ptr->disconnect(quic_id);
    }
    else
    {
        self->mDelayedShutdown = true;
    }
}
```

src/overlay/QUICPeer.h
```C++
#pragma once

#include "overlay/Peer.h"
#include "util/Timer.h"
#include <queue>

#include "overlay/QUIC.h"

namespace medida
{
class Meter;
}

namespace stellar
{

// Peer that communicates via a QUIC socket.
class QUICPeer : public Peer
{
  public:
    typedef asio::buffered_stream<asio::ip::tcp::socket> SocketType;

  private:
    std::vector<uint8_t> mIncomingHeader;
    std::vector<uint8_t> mIncomingBody;

    std::queue<std::shared_ptr<xdr::msg_ptr>> mWriteQueue;
    bool mWriting{false};
    bool mDelayedShutdown{false};
    bool mShutdownScheduled{false};

    PeerBareAddress makeAddress(int remoteListeningPort) const override;
 
    void sendMessage(xdr::msg_ptr&& xdrBytes) override;


    virtual void connected() override;
    void startRead();

    void writeHandler(asio::error_code const& error,
                      std::size_t bytes_transferred) override;
    void readHeaderHandler(asio::error_code const& error,
                           std::size_t bytes_transferred) override;
    void readBodyHandler(asio::error_code const& error,
                         std::size_t bytes_transferred) override;

  public:
    QUIC* quic_ptr;
    int quic_id;

    typedef std::shared_ptr<QUICPeer> pointer;

    QUICPeer(Application& app, Peer::PeerRole role);
    // hollow
    // constuctor; use
    // `initiate` or
    // `accept` instead

    static pointer initiate(Application& app, PeerBareAddress const& address);
    static pointer accept(Application& app, int ID);

    void recvMessage(const char *str);

    virtual ~QUICPeer();

    virtual void drop(bool force = true) override;
};
}
```
