package leaderboard

import (
	"context"

	"github.com/redis/go-redis/v9"
)

const leaderboardKey = "leaderboard:global"

type RedisRepository struct{ client *redis.Client }

func NewRedisRepository(client *redis.Client) *RedisRepository {
	return &RedisRepository{client: client}
}

func (r *RedisRepository) Submit(ctx context.Context, playerID string, score int64) error {
	return r.client.ZAddArgs(ctx, leaderboardKey, redis.ZAddArgs{
		GT:      true,
		Members: []redis.Z{{Score: float64(score), Member: playerID}},
	}).Err()
}

func (r *RedisRepository) Top(ctx context.Context, limit int64) ([]Score, error) {
	values, err := r.client.ZRevRangeWithScores(ctx, leaderboardKey, 0, limit-1).Result()
	if err != nil {
		return nil, err
	}
	scores := make([]Score, 0, len(values))
	for _, value := range values {
		member, ok := value.Member.(string)
		if !ok {
			continue
		}
		scores = append(scores, Score{PlayerID: member, Score: int64(value.Score)})
	}
	return scores, nil
}
