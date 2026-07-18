//go:build windows

package installer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"

	"language-betawi/internal/assets"
	"language-betawi/internal/betawimsg"
)

const installDirName = "Betawi"

func InstallDir() (string, error) {
	base := os.Getenv("ProgramFiles")
	if base == "" {
		return "", &SetupError{Message: betawimsg.InstallProblem("kagak nemu folder Program Files di komputer lu")}
	}
	return filepath.Join(base, installDirName), nil
}

func DetectExistingInstall(installDir string) bool {
	target := filepath.Join(installDir, "betawi.exe")
	info, err := os.Stat(target)
	return err == nil && !info.IsDir()
}

func RunSetup(repairMode bool, report func(progress float64, status string)) error {
	report(0.02, "Ngecek instalasi...")

	installDir, err := InstallDir()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(installDir, 0o755); err != nil {
		return &SetupError{Message: betawimsg.InstallProblem("kagak bisa bikin folder instalasi: " + err.Error())}
	}

	report(0.05, "Ngecek sisa ruang penyimpanan...")
	if err := checkDiskSpace(installDir); err != nil {
		return err
	}

	err = extractCompiler(installDir, repairMode, func(p float64, status string) {
		report(0.10+p*0.75, status)
	})
	if err != nil {
		return err
	}

	report(0.90, "Daftarin Betawi ke System PATH Windows...")
	if err := registerPath(installDir); err != nil {
		return err
	}

	report(1.0, "Kelar!")
	return nil
}

func checkDiskSpace(installDir string) error {
	pathPtr, err := windows.UTF16PtrFromString(installDir)
	if err != nil {
		return &SetupError{Message: betawimsg.InstallProblem(err.Error())}
	}

	var freeBytesAvailable, totalBytes, totalFreeBytes uint64
	if err := windows.GetDiskFreeSpaceEx(pathPtr, &freeBytesAvailable, &totalBytes, &totalFreeBytes); err != nil {

		return nil
	}

	const safetyMarginBytes = 10 * 1024 * 1024
	required := uint64(len(assets.CompilerBinary)) + safetyMarginBytes

	if freeBytesAvailable < required {
		return &SetupError{Message: betawimsg.InsufficientStorage(int64(required), int64(freeBytesAvailable))}
	}
	return nil
}

func extractCompiler(installDir string, repairMode bool, report func(progress float64, status string)) error {
	target := filepath.Join(installDir, "betawi.exe")

	if repairMode {
		if existing, err := os.ReadFile(target); err == nil && bytesEqual(existing, assets.CompilerBinary) {
			report(1.0, "File udah sesuai, kagak perlu ditimpa.")
			return nil
		}
	}

	total := len(assets.CompilerBinary)
	if total == 0 {
		return &SetupError{Message: betawimsg.InstallProblem("betawi.exe yang ke-embed kosong — build.bat kayaknya belom bener")}
	}

	f, err := os.Create(target)
	if err != nil {
		return &SetupError{Message: betawimsg.InstallProblem("gagal bikin betawi.exe: " + err.Error())}
	}
	defer f.Close()

	const chunkSize = 64 * 1024
	written := 0
	for written < total {
		end := written + chunkSize
		if end > total {
			end = total
		}
		n, err := f.Write(assets.CompilerBinary[written:end])
		if err != nil {
			return &SetupError{Message: betawimsg.InstallProblem("gagal nulis betawi.exe: " + err.Error())}
		}
		written += n

		pct := float64(written) / float64(total)
		report(pct, fmt.Sprintf("Ngedumpin betawi.exe... (%d KB / %d KB)", written/1024, total/1024))

		if total < 2*1024*1024 {
			time.Sleep(5 * time.Millisecond)
		}
	}

	if err := os.Chmod(target, 0o755); err != nil {

		_ = err
	}
	return nil
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func registerPath(installDir string) error {
	key, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\Session Manager\Environment`,
		registry.QUERY_VALUE|registry.SET_VALUE,
	)
	if err != nil {
		return &SetupError{Message: betawimsg.InstallProblem(
			"gagal buka registry System PATH (mesti jalan sebagai Administrator): " + err.Error())}
	}
	defer key.Close()

	existing, _, err := key.GetStringValue("Path")
	if err != nil && err != registry.ErrNotExist {
		return &SetupError{Message: betawimsg.InstallProblem("gagal baca PATH yang lama: " + err.Error())}
	}

	if pathContains(existing, installDir) {
		return nil
	}

	newPath := existing
	if newPath != "" && !strings.HasSuffix(newPath, ";") {
		newPath += ";"
	}
	newPath += installDir

	if err := key.SetStringValue("Path", newPath); err != nil {
		return &SetupError{Message: betawimsg.InstallProblem("gagal nyimpen PATH baru: " + err.Error())}
	}

	broadcastEnvironmentChange()
	return nil
}

func pathContains(pathVar, dir string) bool {
	dir = strings.ToLower(strings.TrimRight(dir, `\`))
	for _, entry := range strings.Split(pathVar, ";") {
		if strings.ToLower(strings.TrimRight(strings.TrimSpace(entry), `\`)) == dir {
			return true
		}
	}
	return false
}

func broadcastEnvironmentChange() {
	user32 := syscall.NewLazyDLL("user32.dll")
	sendMessageTimeout := user32.NewProc("SendMessageTimeoutW")

	const (
		hwndBroadcast   = 0xffff
		wmSettingChange = 0x001A
		smtoAbortIfHung = 0x0002
	)

	param, err := syscall.UTF16PtrFromString("Environment")
	if err != nil {
		return
	}

	sendMessageTimeout.Call(
		uintptr(hwndBroadcast),
		uintptr(wmSettingChange),
		0,
		uintptr(unsafe.Pointer(param)),
		uintptr(smtoAbortIfHung),
		uintptr(5000),
		0,
	)
}
