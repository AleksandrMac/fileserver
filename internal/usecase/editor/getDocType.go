package editor_usecase

func getDocType(ext string) string {
	switch ext {
	case ".xls", ".xlsx", ".ods", ".csv", ".ots":
		return "cell"
	case ".ppt", ".pptx", ".odp", ".otp", ".pps":
		return "slide"
	case ".djvu", ".oxps", ".pdf", ".xps":
		return "pdf"
	case ".vsdm", ".vsdx", ".vssm", ".vssx", ".vstm", ".vstx":
		return "diagram"
	default:
		return "word"
	}
}
