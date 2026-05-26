#define MyAppName "Winitrix"
#ifndef MyAppVersion
  #define MyAppVersion "0.0.0"
#endif
#define MyAppPublisher "Winitrix"
#define MyAppURL "https://github.com/thesubh213/winitrix"
#define MyAppExeName "winitrix.exe"

[Setup]
AppId={{E6D8E9E4-6B9E-4F54-9C1C-2D2F9FA3D4B9}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
DefaultDirName={localappdata}\Winitrix
DefaultGroupName=Winitrix
DisableProgramGroupPage=yes
OutputBaseFilename=winitrix-setup-{#MyAppVersion}
OutputDir=..\dist
Compression=lzma
SolidCompression=yes
ChangesEnvironment=yes
PrivilegesRequired=lowest
WizardStyle=modern
LicenseFile=..\LICENSE

[Tasks]
Name: "desktopicon"; Description: "Create a &desktop icon"; Flags: unchecked
Name: "addtopath"; Description: "Add Winitrix to &PATH"; Flags: checkedonce

[Files]
Source: "..\dist\winitrix.exe"; DestDir: "{app}"; DestName: "{#MyAppExeName}"; Flags: ignoreversion
Source: "..\installer\winitrix-launch.cmd"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\README.md"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\LICENSE"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{userprograms}\Winitrix"; Filename: "{app}\winitrix-launch.cmd"; WorkingDir: "{app}"; IconFilename: "{app}\{#MyAppExeName}"
Name: "{userdesktop}\Winitrix"; Filename: "{app}\winitrix-launch.cmd"; WorkingDir: "{app}"; Tasks: desktopicon; IconFilename: "{app}\{#MyAppExeName}"

[Code]
function NormalizePath(Value: string): string;
begin
  Result := Lowercase(Trim(Value));
end;

function NeedsAddPath(Param: string): boolean;
var
  Paths: string;
begin
  if not RegQueryStringValue(HKCU, 'Environment', 'Path', Paths) then
    Paths := '';

  Result := Pos(';' + NormalizePath(Param) + ';', ';' + NormalizePath(Paths) + ';') = 0;
end;

procedure AddToPath(Param: string);
var
  Paths: string;
begin
  if not RegQueryStringValue(HKCU, 'Environment', 'Path', Paths) then
    Paths := '';

  if (Paths <> '') and (Copy(Paths, Length(Paths), 1) <> ';') then
    Paths := Paths + ';';

  Paths := Paths + Param;
  RegWriteStringValue(HKCU, 'Environment', 'Path', Paths);
end;

procedure RemoveFromPath(Param: string);
var
  Paths: string;
  Items: TStringList;
  I: Integer;
  NewPath: string;
begin
  if not RegQueryStringValue(HKCU, 'Environment', 'Path', Paths) then
    exit;

  Items := TStringList.Create;
  try
    Items.Delimiter := ';';
    Items.StrictDelimiter := True;
    Items.DelimitedText := Paths;

    for I := Items.Count - 1 downto 0 do
      if NormalizePath(Items[I]) = NormalizePath(Param) then
        Items.Delete(I);

    NewPath := Items.DelimitedText;
    RegWriteStringValue(HKCU, 'Environment', 'Path', NewPath);
  finally
    Items.Free;
  end;
end;

procedure CurStepChanged(CurStep: TSetupStep);
begin
  if (CurStep = ssPostInstall) and IsTaskSelected('addtopath') then
    if NeedsAddPath(ExpandConstant('{app}')) then
      AddToPath(ExpandConstant('{app}'));
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  if CurUninstallStep = usPostUninstall then
    RemoveFromPath(ExpandConstant('{app}'));
end;
