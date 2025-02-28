package service

type WalletRepository interface {
	Deposit(userID int, amount float64) error
	Transfer(fromUserID, toUserID int, amount float64) error
}

type WalletService struct {
	repo WalletRepository
}

func NewWalletService(repo WalletRepository) *WalletService {
	return &WalletService{repo: repo}
}

func (s *WalletService) Deposit(userID int, amount float64) error {
	return s.repo.Deposit(userID, amount)
}

func (s *WalletService) Transfer(fromUserID, toUserID int, amount float64) error {
	return s.repo.Transfer(fromUserID, toUserID, amount)
}
