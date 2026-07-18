package betawimsg

import "fmt"

func FuzzyWarning(line int, original, matched string, scorePct float64) string {
	return fmt.Sprintf(
		"[Woro-woro] Lu ngetik '%s' di baris %d ye? Gua anggep maksud lu '%s' (cocok %.0f%%). Kode tetep jalan, tapi rapihin lagi besok-besok, tong!",
		original, line, matched, scorePct,
	)
}

func UnknownWord(word string, line int) string {
	return fmt.Sprintf(
		"Waduh amsyong bre! Lu ngetik '%s' di baris %d. Kagak ada kata kayak gitu di kamus kita, betulin sono, tong!",
		word, line,
	)
}

func SyntaxProblem(detail string, line int) string {
	return fmt.Sprintf(
		"Aduh, ada yang kagak beres bre di baris %d: %s. Betulin dulu, baru gas lagi, tong!",
		line, detail,
	)
}

func DBConnectionFailure(detail string) string {
	return fmt.Sprintf(
		"Amsyong bre! Database-nya kaga kesenggol/kagak mau nyolok. Cek kabel atau konfigurasi lu dah! (%s)",
		detail,
	)
}

func RuntimeProblem(detail string, line int) string {
	return fmt.Sprintf(
		"Zonk bre, error di baris %d: %s. Cek lagi kodingannya, tong!",
		line, detail,
	)
}

func ServerCrash(port int64, detail string) string {
	return fmt.Sprintf(
		"Amsyong bre! Warung di port %d kagak bisa dibuka. %s",
		port, detail,
	)
}

func InstallProblem(detail string) string {
	return fmt.Sprintf("Zonk bre, ada masalah pas masang Betawi: %s. Coba lagi, tong!", detail)
}

func InsufficientStorage(requiredBytes, availableBytes int64) string {
	return fmt.Sprintf(
		"Amsyong bre! Penyimpanan lu kagak cukup buat masang Betawi. Butuh sekitar %s, yang kosong cuma %s. Kosongin dulu ruang penyimpanannya, tong!",
		humanizeBytes(requiredBytes), humanizeBytes(availableBytes),
	)
}

func humanizeBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	units := "KMGTPE"
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), units[exp])
}
