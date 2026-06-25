package service

type CrontabService struct{}

func NewCrontabService() *CrontabService {
	return &CrontabService{}
}

func (c *CrontabService) Install(folder string, cronExpr string, configPath string) error {
	return nil
}

func (c *CrontabService) Uninstall() error {
	return nil
}

func (c *CrontabService) Exists() (bool, error) {
	return false, nil
}

func (c *CrontabService) CheckStatus() (string, error) {
	return "", nil
}
