package url

type Url struct {
	Old string
	New string
}
type Urls []Url

func (urls *Urls) Flatten() []string {
	fl := make([]string, len(*urls)*2)
	for _, url := range *urls {
		fl = append(fl, url.Old, url.New)
	}
	return fl
}
