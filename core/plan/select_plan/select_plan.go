/*
Copyright 2018 Sgoby.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreedto in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package select_plan

import (
	"github.com/sgoby/sqlparser"
	"github.com/sgoby/myhub/core/schema"
	"fmt"
	"github.com/sgoby/myhub/core/rule"
	"github.com/sgoby/myhub/core/plan"
	"github.com/sgoby/myhub/core/rule/result"
	"strconv"
)

type selectPlanBuilder struct {
	plan.PlanBuilder
	stmt          *sqlparser.Select
	limitOffset   int64
	limitRowcount int64
}

//
func NewSelectPlanBuilder(stmt sqlparser.Select) (*selectPlanBuilder, error) {
	var offset, rowcount int64
	var err error
	if stmt.Limit != nil {
		if stmt.Limit.Offset != nil {
			tbufOffset := sqlparser.NewTrackedBuffer(nil)
			stmt.Limit.Offset.Format(tbufOffset)
			offset, err = strconv.ParseInt(tbufOffset.String(), 10, 64)
			if err != nil {
				return nil, err
			}
		}
		if stmt.Limit.Rowcount != nil {
			tbufRowcount := sqlparser.NewTrackedBuffer(nil)
			stmt.Limit.Rowcount.Format(tbufRowcount)
			rowcount, err = strconv.ParseInt(tbufRowcount.String(), 10, 64)
			if err != nil {
				return nil, err
			}
		}
	}
	//
	builder := &selectPlanBuilder{
		stmt:          &stmt,
		limitOffset:   offset,
		limitRowcount: rowcount,
	}
	return builder, nil
}

//
func BuildSelectPlan(tb *schema.Table, stmt *sqlparser.Select, manager *rule.RuleManager) ([]plan.Plan, error) {
	if stmt == nil {
		return nil, fmt.Errorf("stmt is nil")
	}
	var offset, rowcount int64
	var err error
	if stmt.Limit != nil {
		if stmt.Limit.Offset != nil {
			tbufOffset := sqlparser.NewTrackedBuffer(nil)
			stmt.Limit.Offset.Format(tbufOffset)
			offset, err = strconv.ParseInt(tbufOffset.String(), 10, 64)
			if err != nil {
				return nil, err
			}
		}
		if stmt.Limit.Rowcount != nil {
			tbufRowcount := sqlparser.NewTrackedBuffer(nil)
			stmt.Limit.Rowcount.Format(tbufRowcount)
			rowcount, err = strconv.ParseInt(tbufRowcount.String(), 10, 64)
			if err != nil {
				return nil, err
			}
		}
	}
	//
	builder := &selectPlanBuilder{
		stmt:          stmt,
		limitOffset:   offset,
		limitRowcount: rowcount,
	}
	//
	expr, isFound := builder.getWhereExprByKey(tb.GetRuleKey())
	if !isFound { //
		expr = nil //get all
	}
	//
	rResults, err := manager.GetShardRule(tb.GetRuleName(), expr)
	if err != nil {
		return nil, err
	}
	//If no matching rule, find all, just for select statement.
	if isFound && (rResults == nil || len(rResults) < 1) {
		rResults, err = manager.GetShardRule(tb.GetRuleName(), nil)
		if err != nil {
			return nil, err
		}
		//default select on first node when not matching rule
		if len(rResults) > 0 {
			rResults = rResults[0:1]
			if len(rResults[0].TbSuffixs) > 0 {
				rResults[0].TbSuffixs = rResults[0].TbSuffixs[0:1]
			}
		}
	}
	return builder.createSelectStmt(rResults, stmt)
}

//
func (this *selectPlanBuilder) createSelectStmt(rResults []result.RuleResult, stmt *sqlparser.Select) ([]plan.Plan, error) {
	var plans []plan.Plan
	for _, rule := range rResults {
		mplan := plan.Plan{
			NodeDBName: rule.NodeDB,
		}
		for _, tbSuffix := range rule.TbSuffixs {
			nStmt := this.tableNameAddSuffix(*stmt, rule.NodeDB, tbSuffix)
			if this.limitRowcount > 0 {
				//change limit rowcount
				nStmt.Limit = new(sqlparser.Limit)
				nStmt.Limit.Offset = sqlparser.NewIntVal([]byte(fmt.Sprintf("%d", 0)))
				nStmt.Limit.Rowcount = sqlparser.NewIntVal([]byte(fmt.Sprintf("%d", this.limitOffset+this.limitRowcount)))
			}
			mplan.AddPlanQuery(&nStmt, "")
		}
		//
		plans = append(plans, mplan)
	}
	return plans, nil
}

//
func (this *selectPlanBuilder) tableNameAddSuffix(stmt sqlparser.Select, dbName, tbSuffix string) sqlparser.Select {
	nStmt := sqlparser.Select{}
	nStmt = stmt
	nStmt.From = this.tableNameAddSuffixFrom(nStmt.From, dbName, tbSuffix)
	return nStmt
}

//
func (this *selectPlanBuilder) tableNameAddSuffixFrom(from sqlparser.TableExprs, dbName, tbSuffix string) sqlparser.TableExprs {
	newFrom := make(sqlparser.TableExprs, 1)
	switch expr := from[0].(type) {
	case *sqlparser.AliasedTableExpr:
		nAli := sqlparser.AliasedTableExpr{
			Partitions: expr.Partitions,
			As:         expr.As,
			Hints:      expr.Hints,
		}
		if tbn, ok := expr.Expr.(sqlparser.TableName); ok {
			oldName := tbn.Name.String()
			newTb := tbn.ToViewName()
			if !tbn.Qualifier.IsEmpty() {
				newTb.Qualifier = sqlparser.NewTableIdent(dbName)
			}
			newTb.Name = sqlparser.NewTableIdent(oldName + "_" + tbSuffix)
			nAli.Expr = newTb
		}
		if tbn, ok := expr.Expr.(*sqlparser.Subquery); ok {
			newTb := new(sqlparser.Subquery)
			newSelect := new(sqlparser.Select)
			if stmt, ok := tbn.Select.(*sqlparser.Select); ok {
				newSelect.Cache = stmt.Cache
				newSelect.Comments = stmt.Comments
				newSelect.Distinct = stmt.Distinct
				newSelect.Hints = stmt.Hints
				newSelect.SelectExprs = stmt.SelectExprs
				//newSelect.From  = stmt.From
				newSelect.Where = stmt.Where
				newSelect.GroupBy = stmt.GroupBy
				newSelect.Having = stmt.Having
				newSelect.OrderBy = stmt.OrderBy
				newSelect.Limit = stmt.Limit
				newSelect.Lock = stmt.Lock
				//
				newSelect.From = this.tableNameAddSuffixFrom(stmt.From, dbName, tbSuffix)
				newTb.Select = newSelect
				nAli.Expr = newTb
			}
		}
		newFrom = make(sqlparser.TableExprs, 1)
		newFrom[0] = &nAli
	case *sqlparser.ParenTableExpr:
		newFrom = from
	case *sqlparser.JoinTableExpr:
		newFrom = from
	default:
		newFrom = from
	}
	return newFrom
}

//
func (this *selectPlanBuilder) getWhereExprByKey(key string) (rExpr sqlparser.Expr, isFound bool) {
	if this.stmt.Where == nil {
		return nil, false
	}
	pExpr := this.stmt.Where.Expr
	return this.GetExprByKey(pExpr, key)
}
