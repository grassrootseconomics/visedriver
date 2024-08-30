package utils




type AccountFileHandlerInterface interface {
    EnsureFileExists() error
    ReadAccountData() (map[string]string, error)
    WriteAccountData(data map[string]string) error
}



