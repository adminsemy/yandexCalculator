package newexpression

import (
	"encoding/json"
	"time"

	"github.com/adminsemy/yandexCalculator/Orchestrator/internal/config"
	"github.com/adminsemy/yandexCalculator/Orchestrator/internal/entity"
	"github.com/adminsemy/yandexCalculator/Orchestrator/internal/storage/memory"
	"github.com/adminsemy/yandexCalculator/Orchestrator/internal/tasks/arithmetic"
	"github.com/adminsemy/yandexCalculator/Orchestrator/internal/tasks/queue"
	"github.com/adminsemy/yandexCalculator/Orchestrator/internal/validator"
)

type NewExpressionAst struct {
	conf    *config.Config
	storage *memory.Storage
	queue   *queue.MapQueue
}

func NewExpression(conf *config.Config, storage *memory.Storage, queue *queue.MapQueue, expression string) ([]byte, error) {
	exp := entity.NewExpression(expression, "", validator.Validator)
	if exp.Err == nil {
		exp.Duration = duration(exp.Expression, conf)
	}
	storage.Set(exp)
	ast, err := arithmetic.NewASTTree(exp, conf, queue)
	if err != nil {
		resp := entity.NewResponseExpression(exp.ID, exp.Expression, time.Now(), 0, false, 0, err)
		data, e := json.Marshal(resp)
		if e != nil {
			return nil, e
		}
		return data, nil
	}
	resp := entity.NewResponseExpression(exp.ID, exp.Expression, ast.Start, exp.Duration, ast.IsCalc, exp.Result, exp.Err)
	data, e := json.Marshal(resp)
	if e != nil {
		return nil, e
	}
	return data, nil
}

func duration(exp string, config *config.Config) int64 {
	res := int64(0)
	for i := 0; i < len(exp); i++ {
		if exp[i] == '+' {
			res += config.Plus
		}
		if exp[i] == '-' {
			if i == 0 ||
				exp[i-1] == '(' ||
				exp[i-1] == '+' ||
				exp[i-1] == '-' ||
				exp[i-1] == '*' ||
				exp[i-1] == '/' {
				continue
			}
			res += config.Minus
		}
		if exp[i] == '*' {
			res += config.Multiply
		}
		if exp[i] == '/' {
			res += config.Divide
		}
	}
	return res
}
