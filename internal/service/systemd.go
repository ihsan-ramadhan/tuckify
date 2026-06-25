package service

type SystemdService struct{}

func NewSystemdService() *SystemdService {
	return &SystemdService{}
}

func (s *SystemdService) Install(folder string, cronExpr string, configPath string) error {
	return nil
}

func (s *SystemdService) Uninstall() error {
	return nil
}

func (s *SystemdService) Exists() (bool, error) {
	return false, nil
}

func (s *SystemdService) CheckStatus() (string, error) {
	return "", nil
}
