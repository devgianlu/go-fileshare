package fileshare

type HttpServer interface {
	ListenForever() error
}
