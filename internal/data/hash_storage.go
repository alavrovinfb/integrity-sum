package data

type HashData struct {
	Hash         string
	FullFileName string
	Algorithm    string
	PodName      string
	ReleaseId    int
}

type HashDataOutput struct {
	ID           int
	Hash         string
	FullFileName string
	Algorithm    string
	NamePod      string
	ReleaseId    int
}
