package container

type ImageManager interface {
	Pull()
	Validate()
}
