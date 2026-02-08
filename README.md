<img src="./src/assets/icons/appicon.png" alt="Winticator logo" title="Winticator" align="left" height="60px" />

# Winticator

[![Go version](https://img.shields.io/github/go-mod/go-version/nktmys/winticator)](https://github.com/nktmys/winticator/blob/main/go.mod) 
[![GitHub release](https://img.shields.io/github/release/nktmys/winticator)](https://github.com/nktmys/winticator/releases/latest) 
[![GitHub all releases](https://img.shields.io/github/downloads/nktmys/winticator/total)](https://github.com/nktmys/winticator/releases) 
[![Build Status](https://img.shields.io/github/actions/workflow/status/nktmys/winticator/test.yaml?label=go%20test)](https://github.com/nktmys/winticator/actions/workflows/test.yaml)

English | [日本語](./README-ja.md)

`Winticator` is an independent, open-source TOTP authenticator for desktop environments
(Windows / macOS / Linux), implemented as a cross-platform reference application.

This project is intended as a **technical demonstration** of how to build and distribute
a standalone, cross-platform desktop application based on open standards.

---

## Overview

- Cross-platform desktop authenticator (Windows / macOS / Linux)
- Based on the open standard **TOTP (RFC 6238)**
- Works entirely offline (no network communication)
- Designed as a simple, minimal implementation
- Released as **MIT licensed open-source software**

Winticator focuses on **technical implementation and portability**,
not on product differentiation or commercial use.

---

## Features

- **QR Code Scanning** — Press the **Add** button to scan TOTP QR codes from your screen automatically
- **Google Authenticator Import** — Bulk import entries via `otpauth-migration://` QR codes
- **Backup & Restore** — Export/import entries using a proprietary password-protected backup file
- **Fully Offline** — No network communication; all data stays on your machine
- **Copy to Clipboard** — Tap an OTP code to copy it instantly
- **Show QR Code** — Display any entry as a QR code for transfer to other devices

---

## Supported URI Schemes

| Scheme | Description |
|--------|-------------|
| `otpauth://` | Standard TOTP URI format (RFC 6238) for adding individual entries |
| `otpauth-migration://` | Google Authenticator export format for bulk import |

> **Note:** `otpauth-migration://` QR codes that contain many entries may be dense and difficult to decode.
> If scanning fails, try changing the display scale and scanning again, or export entries one at a time from the source app.

---

## Screenshots

| TOTP List | Settings |
|:---------:|:--------:|
| <img src="./images/ss-01-totp-list.png" alt="TOTP List" width="350"> | <img src="./images/ss-02-settings.png" alt="Settings" width="350"> |

| App Info | Export |
|:--------:|:------:|
| <img src="./images/ss-03-appinfo.png" alt="App Info" width="350"> | <img src="./images/ss-04-export.png" alt="Export" width="350"> |

---

## Download

Download binaries for your platform from the [Releases](https://github.com/nktmys/winticator/releases/latest) page.

| Platform | Architecture | File |
|----------|--------------|------|
| Windows | x64 (amd64) | `winticator-windows.zip` |
| macOS | Apple Silicon (arm64) | `winticator-macos.zip` |
| Linux | x64 (amd64) | `winticator-linux.tar.xz` |

**If you need binaries for other architectures, feel free to open an [issue](https://github.com/nktmys/winticator/issues).**

---

## Purpose of This Project

The primary purpose of this project is:

- To demonstrate a practical approach for building **cross-platform desktop applications**
- To provide a **reference implementation** of a TOTP authenticator on desktop environments
- To share implementation knowledge using **publicly available standards and information**

This project is **not intended to be a commercial product**,
nor to compete with or replace any existing services or applications.

---

## Standards and References

Winticator is implemented based solely on publicly available materials, including:

- RFC 6238: Time-Based One-Time Password Algorithm
- Public documentation and specifications of TOTP
- Open-source reference implementations

No proprietary algorithms, designs, or internal specifications from any organization
are used in this project.

---

## Non-Affiliation Disclaimer

Winticator is an **independent project**.

- It is **not affiliated with, endorsed by, or associated with Google**
- It is **not affiliated with, endorsed by, or associated with any specific company**
- Any product or service names mentioned are used solely for descriptive purposes

---

## Scope and Limitations

To keep the project intentionally simple and neutral:

- The feature set is minimal
- Advanced usability, enterprise features, or service integrations are out of scope
- No comparison with other authenticator applications is intended

The goal is **clarity and simplicity**, not feature completeness.

---

## License

This project is licensed under the **MIT License**.

You are free to use, modify, distribute, and incorporate this software,
including for commercial purposes, under the terms of the license.

See the [LICENSE](./LICENSE) file for details.

---

## Troubleshooting

### macOS: "App is damaged" or cannot be opened

When launching the app on macOS, you may see an error message such as:
- "Winticator is damaged and can't be opened"
- "Winticator cannot be opened because the developer cannot be verified"

This is caused by macOS Gatekeeper quarantine. To resolve this, run the following command in Terminal:

```bash
xattr -r -d com.apple.quarantine /Applications/Winticator.app
```

Replace `/Applications/Winticator.app` with the actual path where you placed the app.

### macOS: QR code scanning does not work

If QR code scanning does not work even though the "Screen Recording and System Audio Recording" permission is enabled, the permission may not be functioning correctly.

To fix this:
1. Open **System Settings** → **Privacy & Security** → **Screen Recording and System Audio Recording**
2. Remove Winticator from the list by clicking the **−** (minus) button
3. Re-add Winticator and enable the permission

This should restore the screen capture functionality needed for QR code scanning.

---

## Notes

This project is provided **"as is"**, without warranty of any kind.

If you are looking for a production-grade or enterprise-ready solution,
please consider established products or services that provide official support.
