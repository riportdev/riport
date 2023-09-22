<!-- markdownlint-disable -->

## At a glance

<!-- markdownlint-restore -->

Riport helps you to manage your remote servers without the hassle of VPNs, chained SSH connections, jump-hosts, or the
use of commercial tools like TeamViewer and its clones.

Riport acts as server and client establishing permanent or on-demand secure tunnels to devices inside protected intranets
behind a firewall.

All operating systems provide secure and well-established mechanisms for remote management, being SSH and Remote Desktop
the most widely used. Riport makes them accessible easily and securely.

Watch our short [explainer video](https://player.vimeo.com/video/573085727).

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

![Riport Principle](https://raw.githubusercontent.com/realvnc-labs/Riport/master/docs/static/images/Riport-principle.svg 'Riport Principle')

## Instantly launch an Riport server

[![Button Launch Riport  Now](https://img.shields.io/badge/Riport_Server-Launch_Now-brightgreen?style=for-the-badge&logo=Windows%20Terminal)](https://kb.Riport.io/install-the-Riport-server)

🚀 If you are curious and want to try Riport, install your server now in no time. Use our
[server installation script](https://kb.Riport.io/install-the-Riport-server).

## 📖 Documentation

### End-User documentation

[![end-user-documentation](https://img.shields.io/badge/End--User_Documentation-Read_Now-green?style=for-the-badge&logo=Gitbook)](https://kb.Riport.io)

our [end-user documentation](https://kb.Riport.io) with articles on user-friendly installation, settings and secure operation of Riport.

### Technical documentation

[![technical-documentation](https://img.shields.io/badge/Technical_Documentation-Read_Now-orange?style=for-the-badge&logo=Github)](https://oss.Riport.io/)

our [technical documentation](https://oss.Riport.io) with all background information and underlying concepts

## Feedback and Help

**We need your feedback**.
Our vision is to establish Riport as a serious alternative to all the black box software for remote management.
To make it a success, please share your feedback.

### Report bugs

If you encounter errors while installing or using Riport, please let us know.
[File an issue report](https://github.com/realvnc-labs/Riport/issues) here on GitHub.

### Ask question

If you have difficulties installing or using Riport, don't hesitate to ask us anything. Often questions give us a hint
on how to improve the documentation. Do not use issue reports for asking questions.
Use the [discussion forum](https://github.com/realvnc-labs/Riport/discussions) instead.

### Positive Feedback

Please share positive feedback also. Give us a star. Write a review. Share our project page on your social media.
Contribute to the [discussion](https://github.com/realvnc-labs/Riport/discussions). Is Riport suitable for your
needs? What is missing?

### Stay tuned

Click on the Watch button in the top right corner of the [Repository Page](https://github.com/realvnc-labs/Riport).
This way you won't miss any updates and the upcoming features.

We send release notes over our [mailing list](https://subscribe.Riport.io).

## Credits

* Thanks to [jpillora/chisel](https://github.com/jpillora/chisel) for the great groundwork of tunnels
* Image by [pch.vector / Freepik](http://www.freepik.com)

## Versioning model

Riport uses `<major>.<minor>.<buildnumber>` version pattern for compatibility with a maximum number of package managers.

Starting from version 1.0.0 packages with even `<minor>` number are considered stable.
