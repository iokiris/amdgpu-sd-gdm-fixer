

## Auto-healing tool for broken **Xorg** sessions on **GNOME + AMDGPU**

If your AMDGPU driver fails with errors like:

```shell
[drm] psp gfx command LOAD_TA(0x1) failed and response status is (0x7)
[drm] psp gfx command INVOKE_CMD(0x3) failed and response status is (0x4)
amdgpu 0000:07:00.0: amdgpu: Secure display: Generic Failure.
amdgpu 0000:07:00.0: amdgpu: SECUREDISPLAY: query securedisplay TA failed. ret 0x0
```

...and your GNOME won't boot (no desktop, GDM stuck, or blank login) — this tool resets Xorg sessions **without wiping your config**.

---

## ⚠️ Warning

- Doesn't fix kernel-level AMDGPU bugs
- Doesn't work with Wayland
- Doesn't reinstall your GPU drivers

## ✅ What it does

- Checks if Xorg session files are missing:
    - `/usr/share/xsessions/*.desktop`
- Backs up your GNOME settings:
    - `~/.config/dconf/user`
- Reinstalls critical packages:
    - `xserver-xorg-core`
    - `xorg`
    - `x11-common`
    - `gnome-session`
- Restores GNOME config
- Restarts GDM

---


## 🚀 How to use

1. Switch to TTY (Ctrl + Alt + F3 or similar)
2. Login with your user
3. Run:

```bash
sudo go run main.gp
```

---

## 📋 Requirements

- GNOME + GDM
- Xorg (**NOT Wayland!**)
- Debian/Ubuntu-based distro
- sudo access

