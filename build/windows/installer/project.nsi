Unicode true

####
## Please note: Template replacements don't work in this file. They are provided with default defines like
## mentioned underneath.
## If the keyword is not defined, "wails_tools.nsh" will populate them with the values from ProjectInfo.
## If they are defined here, "wails_tools.nsh" will not touch them. This allows to use this project.nsi manually
## from outside of Wails for debugging and development of the installer.
##
## For development first make a wails nsis build to populate the "wails_tools.nsh":
## > wails build --target windows/amd64 --nsis
## Then you can call makensis on this file with specifying the path to your binary:
## For a AMD64 only installer:
## > makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\app.exe
## For a ARM64 only installer:
## > makensis -DARG_WAILS_ARM64_BINARY=..\..\bin\app.exe
## For a installer with both architectures:
## > makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\app-amd64.exe -DARG_WAILS_ARM64_BINARY=..\..\bin\app-arm64.exe
####
## The following information is taken from the ProjectInfo file, but they can be overwritten here.
####
## !define INFO_PROJECTNAME    "MyProject" # Default "{{.Name}}"
## !define INFO_COMPANYNAME    "MyCompany" # Default "{{.Info.CompanyName}}"
## !define INFO_PRODUCTNAME    "MyProduct" # Default "{{.Info.ProductName}}"
!define INFO_PRODUCTVERSION "1.1.0"     # Default "{{.Info.ProductVersion}}"
## !define INFO_COPYRIGHT      "Copyright" # Default "{{.Info.Copyright}}"
###
## !define PRODUCT_EXECUTABLE  "Application.exe"      # Default "${INFO_PROJECTNAME}.exe"
## !define UNINST_KEY_NAME     "UninstKeyInRegistry"  # Default "${INFO_COMPANYNAME}${INFO_PRODUCTNAME}"
####
!define REQUEST_EXECUTION_LEVEL "admin"            # Default "admin"  see also https://nsis.sourceforge.io/Docs/Chapter4.html
####
## Include the wails tools
####
!include "wails_tools.nsh"

# The version information for this two must consist of 4 parts
VIProductVersion "${INFO_PRODUCTVERSION}.0"
VIFileVersion    "${INFO_PRODUCTVERSION}.0"

VIAddVersionKey "CompanyName"     "${INFO_COMPANYNAME}"
VIAddVersionKey "FileDescription" "${INFO_PRODUCTNAME} Installer"
VIAddVersionKey "ProductVersion"  "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion"     "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright"  "${INFO_COPYRIGHT}"
VIAddVersionKey "ProductName"     "${INFO_PRODUCTNAME}"

# Enable HiDPI support. https://nsis.sourceforge.io/Reference/ManifestDPIAware
ManifestDPIAware true

!include "MUI.nsh"
!include "LogicLib.nsh"
!include "nsDialogs.nsh"
!include "WinMessages.nsh"

Var OptionsDialog
Var CheckboxDesktop
Var CheckboxLaunch
Var CheckboxAssoc
Var DesktopShortcut
Var AutoLaunch
Var AssociateFiles

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
!define MUI_HEADERIMAGE
!define MUI_HEADERIMAGE_RIGHT

!define MUI_ABORTWARNING
!define MUI_FINISHPAGE_NOAUTOCLOSE
!define MUI_WELCOMEPAGE_TITLE "Welcome to QuraMate Setup"
!define MUI_WELCOMEPAGE_TEXT "This installer will set up QuraMate on your computer.$\r$\n$\r$\nQuraMate is a lightweight, fast desktop database client to connect, query, and manage with confidence."
!define MUI_DIRECTORYPAGE_TEXT_TOP "Choose where QuraMate should be installed."
!define MUI_FINISHPAGE_TITLE "QuraMate is installed"
!define MUI_FINISHPAGE_TEXT "Setup has finished installing lightweight, fast QuraMate on your computer.$\r$\n$\r$\nClick Finish to exit Setup."

AutoCloseWindow true
!insertmacro MUI_PAGE_WELCOME
Page custom OptionsPageCreate OptionsPageLeave
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_UNPAGE_FINISH

!insertmacro MUI_LANGUAGE "English"

## The following two statements can be used to sign the installer and the uninstaller. The path to the binaries are provided in %1
#!uninstfinalize 'signtool --file "%1"'
#!finalize 'signtool --file "%1"'

Name "${INFO_PRODUCTNAME}"
OutFile "..\..\bin\${INFO_PROJECTNAME}-${ARCH}-installer.exe"
InstallDir "$PROGRAMFILES64\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"
ShowInstDetails show

Function .onInit
    !insertmacro wails.checkArchitecture
    StrCpy $DesktopShortcut 1
    StrCpy $AutoLaunch 1
    StrCpy $AssociateFiles 0
FunctionEnd

Function OptionsPageCreate
    nsDialogs::Create 1018
    Pop $OptionsDialog

    ${If} $OptionsDialog == error
        Abort
    ${EndIf}

    ${NSD_CreateLabel} 0 0 100% 24u "Quick setup: choose optional features for a lightweight and fast install."
    Pop $0

    ${NSD_CreateCheckbox} 0 36u 100% 10u "Create Desktop shortcut"
    Pop $CheckboxDesktop
    ${If} $DesktopShortcut == 1
        ${NSD_Check} $CheckboxDesktop
    ${EndIf}

    ${NSD_CreateCheckbox} 0 56u 100% 10u "Launch QuraMate after installation"
    Pop $CheckboxLaunch
    ${If} $AutoLaunch == 1
        ${NSD_Check} $CheckboxLaunch
    ${EndIf}

    ${NSD_CreateCheckbox} 0 76u 100% 10u "Enable file/protocol associations"
    Pop $CheckboxAssoc
    ${If} $AssociateFiles == 1
        ${NSD_Check} $CheckboxAssoc
    ${EndIf}

    nsDialogs::Show
FunctionEnd

Function OptionsPageLeave
    ${NSD_GetState} $CheckboxDesktop $0
    ${If} $0 == ${BST_CHECKED}
        StrCpy $DesktopShortcut 1
    ${Else}
        StrCpy $DesktopShortcut 0
    ${EndIf}

    ${NSD_GetState} $CheckboxLaunch $0
    ${If} $0 == ${BST_CHECKED}
        StrCpy $AutoLaunch 1
    ${Else}
        StrCpy $AutoLaunch 0
    ${EndIf}

    ${NSD_GetState} $CheckboxAssoc $0
    ${If} $0 == ${BST_CHECKED}
        StrCpy $AssociateFiles 1
    ${Else}
        StrCpy $AssociateFiles 0
    ${EndIf}
FunctionEnd

Section
    !insertmacro wails.setShellContext

    !insertmacro wails.webview2runtime

    SetOutPath $INSTDIR

    !insertmacro wails.files

    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"

    ${If} $DesktopShortcut == 1
        CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    ${EndIf}

    ${If} $AssociateFiles == 1
        !insertmacro wails.associateFiles
        !insertmacro wails.associateCustomProtocols
    ${EndIf}

    !insertmacro wails.writeUninstaller
SectionEnd

Function .onInstSuccess
    ${If} $AutoLaunch == 1
        Exec '"$INSTDIR\${PRODUCT_EXECUTABLE}"'
    ${EndIf}
FunctionEnd

Section "uninstall"
    !insertmacro wails.setShellContext

    RMDir /r "$AppData\${PRODUCT_EXECUTABLE}"

    RMDir /r $INSTDIR

    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    !insertmacro wails.unassociateFiles
    !insertmacro wails.unassociateCustomProtocols

    !insertmacro wails.deleteUninstaller
SectionEnd
