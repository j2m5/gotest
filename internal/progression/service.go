package progression

import "context"

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) Grant(ctx context.Context, reward Reward) (*Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := ValidateReward(reward); err != nil {
		return nil, err
	}
	return s.repository.Grant(ctx, reward)
}

// ValidateReward enforces the same key format as the persistent reward ledger:
// 1..128 ASCII bytes, starting with a letter/digit, then letters/digits or :._/-.
func ValidateReward(reward Reward) error {
	if reward.CharacterID <= 0 || reward.Experience <= 0 || len(reward.Key) == 0 || len(reward.Key) > 128 {
		return ErrInvalidReward
	}
	for i := 0; i < len(reward.Key); i++ {
		b := reward.Key[i]
		alphanumeric := b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9'
		separator := b == ':' || b == '.' || b == '_' || b == '/' || b == '-'
		if !alphanumeric && (i == 0 || !separator) {
			return ErrInvalidReward
		}
	}
	return nil
}
