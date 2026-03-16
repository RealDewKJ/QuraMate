# QuraMate Windows Bootstrapper

This app is a dedicated Windows setup shell for QuraMate.

It exists to provide a branded installation and update experience that is not constrained by the default NSIS wizard UI.

## Responsibilities

- Render the custom QuraMate setup interface
- Resolve the Windows installer path
- Hand off to the silent installer in the background
- Reopen `QuraMate.exe` after installation when possible

## Local build

```powershell
cd bootstrapper/windows/frontend
npm install
npm run build

cd ..
wails build -platform windows/amd64 -nopackage -o QuraMate-Setup.exe
```

## Release flow

The GitHub release workflow now builds this bootstrapper alongside the standard Windows installer and publishes both executables as Windows release assets.
