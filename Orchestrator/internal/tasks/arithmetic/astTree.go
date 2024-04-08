package arithmetic

import (
	"encoding/json"
	"errors"
	"go/ast"
	"go/parser"
	"log/slog"
	"strconv"
	"sync"

	"github.com/adminsemy/yandexCalculator/Orchestrator/internal/config"
	"github.com/adminsemy/yandexCalculator/Orchestrator/internal/entity"
	"github.com/adminsemy/yandexCalculator/Orchestrator/internal/tasks/queue"
	"github.com/adminsemy/yandexCalculator/Orchestrator/internal/tasks/upgrade"
	"github.com/adminsemy/yandexCalculator/Orchestrator/internal/web_socket/client"
)

type ASTTree struct {
	expression *entity.Expression
	X          *ASTTree
	Y          *ASTTree
	Operator   string
	Value      float64
	IsCalc     bool
	IsParent   bool
	queue      *queue.MapQueue
	config     *config.Config
	Err        error
	sync.Mutex
}

type result struct {
	err error
	res float64
}

// Создаем AST дерево из выражения
func NewASTTree(expression *entity.Expression,
	config *config.Config,
	queue *queue.MapQueue,
) (*ASTTree, error) {
	if expression.Err != nil {
		return nil, expression.Err
	}
	// Добавляем кавычки, где только возможно
	upgradeExp := upgrade.Upgrade([]byte(expression.Expression))
	slog.Info("Усовершенствованное выражение", "выражение:", string(upgradeExp))
	// Создаем AST дерево
	tr, err := parser.ParseExpr(string(upgradeExp))
	if err != nil {
		return nil, err
	}
	// Преобразуем AST дерево в нашу структуру ASTTree
	a := create(tr)
	a.expression = expression
	a.queue = queue
	a.config = config
	a.IsCalc = expression.IsCalc
	go a.calc()
	return a, nil
}

// Создаем AST дерево из базы данных
func NewASTTreeDB(
	id uint64,
	expression *entity.Expression,
	value string,
	isErr bool,
	currentResult string,
	config *config.Config,
	queue *queue.MapQueue,
) (*ASTTree, error) {
	tr, err := parser.ParseExpr(currentResult)
	if err != nil {
		return nil, err
	}
	a := create(tr)
	a.expression = expression
	a.Value, _ = strconv.ParseFloat(value, 64)
	if isErr {
		a.Err = errors.New("error")
	}
	a.IsCalc = true
	a.queue = queue
	a.config = config
	return a, nil
}

func create(tr ast.Expr) *ASTTree {
	a := new(ASTTree)
	switch nod := tr.(type) {
	case *ast.BasicLit:
		a.Value, _ = strconv.ParseFloat(nod.Value, 64)
		a.IsCalc = true
	case *ast.ParenExpr:
		a.X = create(nod.X)
		a.IsParent = true
	case *ast.BinaryExpr:
		a.X = create(nod.X)
		a.Y = create(nod.Y)
		a.Operator = nod.Op.String()
	case *ast.UnaryExpr:
		v := create(nod.X).Value
		if nod.Op.String() == "-" {
			v = -v
		}
		a.Value = v
		a.IsCalc = true
	}
	return a
}

// Вычисляем выражение
func (a *ASTTree) calc() {
	if a.IsCalc || a.Err != nil || a == nil {
		return
	}
	var err error
	ch := make(chan result)
	go getResult(a, ch, a, "P")
	res := <-ch
	a.Lock()
	if res.err != nil {
		a.Err = res.err
		a.expression.Err = res.err
	} else {
		a.Value = res.res
		a.IsCalc = true
		a.expression.Result = res.res
		if res.err != nil {
			a.expression.Err = err
		}
		a.expression.IsCalc = true
	}
	a.Unlock()
	resp := entity.NewResponseExpression(a.expression.ID, a.expression.Expression, a.expression.Start, a.expression.Duration, a.IsCalc, a.expression.Result, a.Err)
	answer, err := json.Marshal(resp)
	if err != nil {
		slog.Error("Проблема с формированием ответа", "ошибка:", err)
		return
	}
	go func() {
		a.config.WSmanager.MessageCh <- &client.Message{
			Payload: answer,
			Type:    client.ClientExpression,
		}
	}()
	slog.Info("Выражение вычислено", "выражение:", a.expression.Expression, "результат:", a.expression.Result)
}

// Печатаем полученное выражение, вычисленное в процессе, что бы
// не считать все выражение заново
func PrintExpression(a *ASTTree) string {
	if a.IsCalc {
		return strconv.FormatFloat(a.Value, 'f', -1, 64)
	}
	if a.IsParent {
		return "(" + PrintExpression(a.X) + ")"
	}

	return PrintExpression(a.X) + a.Operator + PrintExpression(a.Y)
}

// Вычисляем каждую операцию. Если уже вычислено
// то возвращаем результат
func getResult(a *ASTTree, ch chan result, parent *ASTTree, level string) {
	if a == nil {
		return
	}
	res := result{}
	if a.IsCalc {
		res.res = a.Value
		ch <- res
		return
	}
	resChX := make(chan result)
	go getResult(a.X, resChX, parent, level+"X")
	if a.Y == nil {
		res = <-resChX
		if res.err != nil {
			ch <- res
			return
		}
		a.Lock()
		a.Value = res.res
		a.IsCalc = true
		a.Unlock()
		ch <- res
		return
	}
	resChY := make(chan result)
	go getResult(a.Y, resChY, parent, level+"Y")
	resX := <-resChX
	if resX.err != nil {
		ch <- resX
		return
	}
	resY := <-resChY
	if resY.err != nil {
		ch <- resY
		return
	}
	res = calculate(resX.res, a.Operator, resY.res, parent, level)
	ch <- res
	if res.err != nil {
		return
	}
	a.Lock()
	a.Value = res.res
	a.IsCalc = true
	a.Unlock()
}

// Вычисляем операцию в зависимости от оператора
func calculate(resX float64, operator string, resY float64, parent *ASTTree, level string) result {
	send := entity.NewOperation(parent.expression.Expression+"-"+level, resX, resY, operator)
	parent.queue.Enqueue(send)
	res := result{}
	resExp := <-send.ResultChan()
	if send.GetError() != nil {
		res.err = send.GetError()
		return res
	}
	res.res = resExp
	return res
}
