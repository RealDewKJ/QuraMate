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
Var SilentAutoLaunch
Var WelcomeDialog
Var FontHero
Var FontTitle
Var FontBody
Var FontCaption
Var FontOption
Var FontMetric
Var WelcomeIconHandle
Var OptionsIconHandle

!define BRAND_ACCENT "E8681D"
!define BRAND_ACCENT_SOFT "FFF4EC"
!define BRAND_TEXT "1A1A1A"
!define BRAND_MUTED "666666"
!define BRAND_PANEL "F6F2ED"
!define BRAND_LINE "E8DED4"
!define BRAND_SURFACE "FFFDFC"
!define BRAND_CARD "FFFFFF"
!define BRAND_CARD_SOFT "FBF2EB"

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"

!define MUI_ABORTWARNING
!define MUI_FINISHPAGE_NOAUTOCLOSE
!define MUI_DIRECTORYPAGE_TEXT_TOP "Choose where QuraMate should be installed."
!define MUI_FINISHPAGE_TITLE "QuraMate is installed"
!define MUI_FINISHPAGE_TEXT "Setup has finished installing lightweight, fast QuraMate on your computer.$\r$\n$\r$\nClick Finish to exit Setup."

AutoCloseWindow true
BrandingText " "
Page custom WelcomePageCreate WelcomePageLeave
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
InstallDirRegKey HKLM "${UNINST_KEY}" "InstallLocation"
ShowInstDetails show

Function .onInit
    !insertmacro wails.checkArchitecture
    StrCpy $DesktopShortcut 1
    StrCpy $AutoLaunch 1
    StrCpy $AssociateFiles 0
    StrCpy $SilentAutoLaunch ""
    CreateFont $FontHero "Segoe UI Semibold" 20 700
    CreateFont $FontTitle "Segoe UI Semibold" 12 700
    CreateFont $FontBody "Segoe UI" 9 400
    CreateFont $FontCaption "Segoe UI Semibold" 8 500
    CreateFont $FontOption "Segoe UI Semibold" 9 600
    CreateFont $FontMetric "Segoe UI Semibold" 16 700

    ${GetParameters} $0
    ${GetOptions} $0 "/AUTOLAUNCHAPP=" $SilentAutoLaunch

    IfSilent 0 done
        StrCpy $DesktopShortcut 0
        StrCpy $AutoLaunch 1
        StrCpy $AssociateFiles 0
    done:
FunctionEnd

Function WelcomePageCreate
    nsDialogs::Create 1018
    Pop $WelcomeDialog

    ${If} $WelcomeDialog == error
        Abort
    ${EndIf}

    ${NSD_CreateLabel} 0 0 100% 100% ""
    Pop $0
    SetCtlColors $0 "" "${BRAND_SURFACE}"

    ${NSD_CreateLabel} 0 0 100% 56u ""
    Pop $1
    SetCtlColors $1 "" "${BRAND_CARD_SOFT}"

    ${NSD_CreateLabel} 12u 10u 60% 10u "SMART DATABASE WORKFLOW"
    Pop $2
    SendMessage $2 ${WM_SETFONT} $FontCaption 1
    SetCtlColors $2 "${BRAND_ACCENT}" "${BRAND_CARD_SOFT}"

    ${NSD_CreateLabel} 12u 20u 60% 18u "QuraMate Setup"
    Pop $3
    SendMessage $3 ${WM_SETFONT} $FontHero 1
    SetCtlColors $3 "${BRAND_TEXT}" "${BRAND_CARD_SOFT}"

    ${NSD_CreateLabel} 12u 40u 60% 12u "Install a focused database workspace with a polished feel, fast launch, and background updates that stay out of the way."
    Pop $4
    SendMessage $4 ${WM_SETFONT} $FontBody 1
    SetCtlColors $4 "${BRAND_MUTED}" "${BRAND_CARD_SOFT}"

    ${NSD_CreateLabel} 74% 8u 20% 40u ""
    Pop $5
    SetCtlColors $5 "" "${BRAND_ACCENT}"

    ${NSD_CreateIcon} 77% 12u 14% 32u ""
    Pop $6
    ${NSD_SetIconFromInstaller} $6 $WelcomeIconHandle

    ${NSD_CreateLabel} 12u 68u 100% 10u "WHAT YOU GET"
    Pop $7
    SendMessage $7 ${WM_SETFONT} $FontCaption 1
    SetCtlColors $7 "${BRAND_ACCENT}" "${BRAND_SURFACE}"

    ${NSD_CreateLabel} 12u 82u 31% 62u ""
    Pop $8
    SetCtlColors $8 "" "${BRAND_CARD}"

    ${NSD_CreateLabel} 18u 90u 22% 10u "01"
    Pop $9
    SendMessage $9 ${WM_SETFONT} $FontMetric 1
    SetCtlColors $9 "${BRAND_ACCENT}" "${BRAND_CARD}"

    ${NSD_CreateLabel} 18u 106u 22% 10u "Clean install"
    Pop $R0
    SendMessage $R0 ${WM_SETFONT} $FontTitle 1
    SetCtlColors $R0 "${BRAND_TEXT}" "${BRAND_CARD}"

    ${NSD_CreateLabel} 18u 120u 22% 18u "Installs into Program Files with a minimal desktop footprint."
    Pop $R1
    SendMessage $R1 ${WM_SETFONT} $FontBody 1
    SetCtlColors $R1 "${BRAND_MUTED}" "${BRAND_CARD}"

    ${NSD_CreateLabel} 35% 82u 31% 62u ""
    Pop $R2
    SetCtlColors $R2 "" "${BRAND_CARD}"

    ${NSD_CreateLabel} 38% 90u 22% 10u "02"
    Pop $R3
    SendMessage $R3 ${WM_SETFONT} $FontMetric 1
    SetCtlColors $R3 "${BRAND_ACCENT}" "${BRAND_CARD}"

    ${NSD_CreateLabel} 38% 106u 22% 10u "Tailored launch"
    Pop $R4
    SendMessage $R4 ${WM_SETFONT} $FontTitle 1
    SetCtlColors $R4 "${BRAND_TEXT}" "${BRAND_CARD}"

    ${NSD_CreateLabel} 38% 120u 22% 18u "Choose shortcuts and launch behavior that fits your workflow."
    Pop $R5
    SendMessage $R5 ${WM_SETFONT} $FontBody 1
    SetCtlColors $R5 "${BRAND_MUTED}" "${BRAND_CARD}"

    ${NSD_CreateLabel} 68% 82u 31% 62u ""
    Pop $R6
    SetCtlColors $R6 "" "${BRAND_CARD_SOFT}"

    ${NSD_CreateLabel} 71% 90u 22% 10u "03"
    Pop $R7
    SendMessage $R7 ${WM_SETFONT} $FontMetric 1
    SetCtlColors $R7 "${BRAND_ACCENT}" "${BRAND_CARD_SOFT}"

    ${NSD_CreateLabel} 71% 106u 22% 10u "Silent updates"
    Pop $R8
    SendMessage $R8 ${WM_SETFONT} $FontTitle 1
    SetCtlColors $R8 "${BRAND_TEXT}" "${BRAND_CARD_SOFT}"

    ${NSD_CreateLabel} 71% 120u 22% 18u "Future updates can install quietly and reopen QuraMate automatically."
    Pop $R9
    SendMessage $R9 ${WM_SETFONT} $FontBody 1
    SetCtlColors $R9 "${BRAND_MUTED}" "${BRAND_CARD_SOFT}"

    nsDialogs::Show

    ${NSD_FreeIcon} $WelcomeIconHandle
FunctionEnd

Function WelcomePageLeave
FunctionEnd

Function OptionsPageCreate
    nsDialogs::Create 1018
    Pop $OptionsDialog

    ${If} $OptionsDialog == error
        Abort
    ${EndIf}

    ${NSD_CreateLabel} 0 0 100% 100% ""
    Pop $0
    SetCtlColors $0 "" "${BRAND_SURFACE}"

    ${NSD_CreateLabel} 0 0 100% 52u ""
    Pop $1
    SetCtlColors $1 "" "${BRAND_CARD_SOFT}"

    ${NSD_CreateLabel} 12u 8u 58% 10u "EXPERIENCE SETTINGS"
    Pop $2
    SendMessage $2 ${WM_SETFONT} $FontCaption 1
    SetCtlColors $2 "${BRAND_ACCENT}" "${BRAND_CARD_SOFT}"

    ${NSD_CreateLabel} 12u 18u 58% 16u "Shape the first-run experience"
    Pop $3
    SendMessage $3 ${WM_SETFONT} $FontHero 1
    SetCtlColors $3 "${BRAND_TEXT}" "${BRAND_CARD_SOFT}"

    ${NSD_CreateLabel} 12u 36u 58% 10u "These choices only affect convenience features. QuraMate itself installs the same either way."
    Pop $4
    SendMessage $4 ${WM_SETFONT} $FontBody 1
    SetCtlColors $4 "${BRAND_MUTED}" "${BRAND_CARD_SOFT}"

    ${NSD_CreateIcon} 82% 10u 10% 28u ""
    Pop $5
    ${NSD_SetIconFromInstaller} $5 $OptionsIconHandle

    ${NSD_CreateCheckbox} 10u 68u 88% 10u "Create Desktop shortcut"
    Pop $CheckboxDesktop
    SendMessage $CheckboxDesktop ${WM_SETFONT} $FontOption 1
    ${If} $DesktopShortcut == 1
        ${NSD_Check} $CheckboxDesktop
    ${EndIf}

    ${NSD_CreateLabel} 10u 84u 88% 18u ""
    Pop $6
    SetCtlColors $6 "" "${BRAND_CARD}"

    ${NSD_CreateLabel} 18u 89u 78% 10u "Keep QuraMate within reach from the desktop for one-click access."
    Pop $7
    SendMessage $7 ${WM_SETFONT} $FontBody 1
    SetCtlColors $7 "${BRAND_MUTED}" "${BRAND_CARD}"

    ${NSD_CreateCheckbox} 10u 110u 88% 10u "Launch QuraMate after installation"
    Pop $CheckboxLaunch
    SendMessage $CheckboxLaunch ${WM_SETFONT} $FontOption 1
    ${If} $AutoLaunch == 1
        ${NSD_Check} $CheckboxLaunch
    ${EndIf}

    ${NSD_CreateLabel} 10u 126u 88% 18u ""
    Pop $8
    SetCtlColors $8 "" "${BRAND_CARD}"

    ${NSD_CreateLabel} 18u 131u 78% 10u "Recommended if you want the app ready the moment setup finishes."
    Pop $9
    SendMessage $9 ${WM_SETFONT} $FontBody 1
    SetCtlColors $9 "${BRAND_MUTED}" "${BRAND_CARD}"

    ${NSD_CreateCheckbox} 10u 152u 88% 10u "Enable file/protocol associations"
    Pop $CheckboxAssoc
    SendMessage $CheckboxAssoc ${WM_SETFONT} $FontOption 1
    ${If} $AssociateFiles == 1
        ${NSD_Check} $CheckboxAssoc
    ${EndIf}

    ${NSD_CreateLabel} 10u 168u 88% 20u ""
    Pop $R0
    SetCtlColors $R0 "" "${BRAND_CARD_SOFT}"

    ${NSD_CreateLabel} 18u 173u 78% 12u "Optional: only turn this on if you want links or related files to open directly with QuraMate."
    Pop $R1
    SendMessage $R1 ${WM_SETFONT} $FontBody 1
    SetCtlColors $R1 "${BRAND_MUTED}" "${BRAND_CARD_SOFT}"

    nsDialogs::Show

    ${NSD_FreeIcon} $OptionsIconHandle
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
    WriteRegStr HKLM "${UNINST_KEY}" "InstallLocation" "$INSTDIR"
SectionEnd

Function .onInstSuccess
    IfSilent 0 interactiveLaunch
        Return

    interactiveLaunch:
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
