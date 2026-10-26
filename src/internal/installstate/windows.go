//go:build windows

package installstate

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode/utf16"

	"sgl.local/sistema-gestion-liderazgo/internal/processutil"
)

func quoted(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }

func powershell(script string) error {
	units := utf16.Encode([]rune("$ErrorActionPreference='Stop';" + script))
	b := make([]byte, len(units)*2)
	for i, u := range units {
		binary.LittleEndian.PutUint16(b[i*2:], u)
	}
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-EncodedCommand", base64.StdEncoding.EncodeToString(b))
	processutil.Hidden(cmd)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func registryPath(root string) string {
	hash := sha256.Sum256([]byte(strings.ToLower(root)))
	return fmt.Sprintf(`HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\MinisterioMusicaOnline-%x`, hash[:8])
}

// Registration is per user; an ordinary installation never needs elevation.
func Register(root, version, icon string) error {
	return powershell(`$root=` + quoted(root) + `;$key=` + quoted(registryPath(root)) + `;
$w=New-Object -ComObject WScript.Shell;
$programs=[IO.Path]::Combine([Environment]::GetFolderPath('StartMenu'),'Programs');
foreach($base in @([Environment]::GetFolderPath('Desktop'),$programs)){
 $s=$w.CreateShortcut([IO.Path]::Combine($base,'Ministerio de la Musica.lnk'));
 $s.TargetPath=[IO.Path]::Combine($root,'Musica.exe');$s.WorkingDirectory=$root;$s.IconLocation=` + quoted(icon) + `;$s.Description='Ministerio de la Música';$s.Save()
};
$s=$w.CreateShortcut([IO.Path]::Combine($programs,'Desinstalar Ministerio de la Musica.lnk'));
$s.TargetPath=[IO.Path]::Combine($root,'` + UninstallerName + `');$s.WorkingDirectory=$root;$s.IconLocation=` + quoted(icon) + `;$s.Description='Desinstalar conservando los datos';$s.Save();
$null=New-Item -Path $key -Force;
(@{DisplayName='Ministerio de la Música Online';DisplayVersion=` + quoted(strings.TrimSuffix(version, "-musica-online")) + `;Publisher='Misión Boliviana Occidental Sur';InstallLocation=$root;DisplayIcon=` + quoted(icon) + `;UninstallString=('"'+[IO.Path]::Combine($root,'` + UninstallerName + `')+'"')}).GetEnumerator() | ForEach-Object { $null=New-ItemProperty -Path $key -Name $_.Key -Value $_.Value -PropertyType String -Force };
$null=New-ItemProperty -Path $key -Name NoModify -Value 1 -PropertyType DWord -Force;
$null=New-ItemProperty -Path $key -Name NoRepair -Value 1 -PropertyType DWord -Force;
$null=Add-Type -MemberDefinition '[DllImport("shell32.dll")] public static extern void SHChangeNotify(uint e,uint f,IntPtr a,IntPtr b);' -Name IconRefresh625 -Namespace MinisterioMusica -PassThru;
[MinisterioMusica.IconRefresh625]::SHChangeNotify(0x08000000,0,[IntPtr]::Zero,[IntPtr]::Zero)`)
}

// Only the optional startup task pointing to this installation may be removed.
// If it runs as SYSTEM, Windows requires the user to run this uninstaller as
// administrator. Failure aborts before removing any application files.
func RemoveStartupTask(root string) error {
	return powershell(`$exe=` + quoted(filepath.Join(root, "ServidorMusica.exe")) + `;
$task=Get-ScheduledTask | Where-Object {$_.TaskName -eq 'MinisterioMusicaOnline' -and $_.TaskPath -eq '\'};
if($task){foreach($a in $task.Actions){if($a.Execute.Trim('"') -ieq $exe){
 try {Unregister-ScheduledTask -InputObject $task -Confirm:$false} catch {throw 'El arranque automático requiere permisos de administrador. Cierre esta ventana y ejecute DESINSTALAR_MinisterioMusica.exe con Ejecutar como administrador.'};break
}}}`)
}

func Unregister(root string) error {
	return powershell(`$root=` + quoted(root) + `;$key=` + quoted(registryPath(root)) + `;
$w=New-Object -ComObject WScript.Shell;
$programs=[IO.Path]::Combine([Environment]::GetFolderPath('StartMenu'),'Programs');
foreach($pair in @(@([Environment]::GetFolderPath('Desktop'),'Ministerio de la Musica.lnk','Musica.exe'),@($programs,'Ministerio de la Musica.lnk','Musica.exe'),@($programs,'Desinstalar Ministerio de la Musica.lnk','` + UninstallerName + `'))){
 $p=[IO.Path]::Combine($pair[0],$pair[1]);if(Test-Path -LiteralPath $p){$s=$w.CreateShortcut($p);if($s.TargetPath -ieq [IO.Path]::Combine($root,$pair[2])){Remove-Item -LiteralPath $p -Force}}
};
if(Test-Path -LiteralPath $key){if((Get-ItemProperty -LiteralPath $key).InstallLocation -ieq $root){Remove-Item -LiteralPath $key -Recurse -Force}}`)
}
