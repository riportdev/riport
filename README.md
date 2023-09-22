<!-- markdownlint-disable -->

## At a glance

<!-- markdownlint-restore -->

Riport helps you to manage your remote servers without the hassle of VPNs, chained SSH connections, jump-hosts, or the
use of commercial tools like TeamViewer and its clones.

Riport acts as server and client establishing permanent or on-demand secure tunnels to devices inside protected intranets
behind a firewall.

All operating systems provide secure and well-established mechanisms for remote management, being SSH and Remote Desktop
the most widely used. Riport makes them accessible easily and securely.

**Is Riport a replacement for TeamViewer?**
Yes and no. It depends on your needs.
TeamViewer and a couple of similar products are focused on giving access to a remote graphical desktop bypassing the
Remote Desktop implementation of Microsoft. They fall short in a heterogeneous environment where access to headless
Linux machines is needed. But they are without alternatives for Windows Home Editions.
Apart from remote management, they offer supplementary services like Video Conferences, desktop sharing, screen
mirroring, or spontaneous remote assistance for desktop users.

**Goal of Riport**
Riport focuses only on remote management of those operating systems where an existing login mechanism can be used.
It can be used for Linux and Windows, but also appliances and IoT devices providing a web-based configuration.
From a technological perspective, [Ngrok](https://ngrok.com/) and [openport.io](https://openport.io) are similar
products. Riport differs from them in many aspects.

* Riport is 100% open source. Client and Server. Remote management is a matter of trust and security. Riport is fully transparent.
* Riport will come with a user interface making the management of remote systems easy and user-friendly.
* Riport is made for all operating systems with native and small binaries. No need for Python or similar heavyweights.
* Riport allows you to self-host the server.
* Riport allows clients to wait in standby mode without an active tunnel. Tunnels can be requested on-demand by the user remotely.

**Supported operating systems**
For the client almost all operating systems are supported, and we provide binaries for a variety of Linux architectures
and Microsoft Windows.
Also, the server can run on any operating system supported by the golang compiler. At the moment we provide server
binaries only for Linux X64 because this is the ideal platform for running it securely and cost-effective.
