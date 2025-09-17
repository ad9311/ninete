package logic

import "context"

func (s *Store) SignOutUser(ctx context.Context, tokenStr string) error {
	tokenHash := hashToken(tokenStr)

	_, err := s.queries.DeleteRefreshToken(ctx, tokenHash)
	if err != nil {
		return HandleDBError(err)
	}

	return nil
}
