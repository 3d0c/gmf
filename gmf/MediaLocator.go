package gmf

import "strings"

type MediaLocator struct {
	Filename string //The Filename which the Medialocator points to, this can also be an URL
	Format   string //forces to use the Fileformat, normaly it will be guessed from the Filename
}

func (loc *MediaLocator) GetProtocol() string {
	lines := strings.Split(loc.Filename, ":")
	if len(lines) != 2 {
		return "file"
	}
	return lines[0]
}

func (loc *MediaLocator) GetReminder() string {
	lines := strings.Split(loc.Filename, ":")
	if len(lines) != 2 {
		return loc.Filename
	}
	return lines[1][2:]
}
