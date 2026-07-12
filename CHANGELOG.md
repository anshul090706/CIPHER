# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial production-grade repository structure setup (docs, scripts, Makefile, etc.).
- Provider startup now logs the libp2p peer ID and full provider multiaddrs for easier client connection setup.
- Added a `-quic` CLI flag for both provider and client to enable QUIC transport explicitly.QUIC is not working on mine. 
- Added a shared 30-second P2P operation timeout for host startup, peer connection, chunk requests, mDNS peer connects, and stream handshakes.It was done becuz in my phase2 task i did the same.

### Changed
- QUIC transport is now disabled by default; TCP/WebSocket remain enabled for stable local runs and tests.
- P2P tests now use bounded contexts so failed networking operations do not hang indefinitely.

### Fixed
- Fixed local P2P verification by making provider connection details visible in provider logs.
- Fixed the default P2P test path by avoiding the unstable QUIC dependency path unless `-quic` is requested.
