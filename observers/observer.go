package observers

type Observer interface {
	Open()
	Exec()
	Close()
}
