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

package stmt

import(
	"github.com/sgoby/sqlparser"
	"github.com/sgoby/myhub/config"
	"github.com/sgoby/myhub/core/rule/result"
	"strings"
	"strconv"
	"fmt"
	"hash/crc32"
)
//config ex:<shard nodeDataBase="db1" between="0-10"/>
type RuleHash struct {
	RuleRange
	maxLen int
}
func NewRuleHash(rcnf config.Rule) (*RuleHash,error){
	rHash := &RuleHash{
		maxLen:rcnf.MaxLen,
	}
	//
	if rHash.maxLen <= 0{
		rHash.maxLen = 1024
	}
	rHash.rangeType = RANGE_NUMERIE
	beginVal := int64(0)
	rHash.strconvInt64 = rHash.strconvInt64Entity
	for _,shcnf := range rcnf.Shards{
		if len(shcnf.RowLimit) < 1{
			shcnf.RowLimit = "1"
		}
		//
		rowLimit, err := strconv.ParseInt(shcnf.RowLimit, 10, 64)
		if err != nil{
			return nil, err
		}
		//optimization 'RangeExpr'
		if rowLimit > 1 {
			begin, end, err := rHash.parserRangeExpr(shcnf.RangeExpr)
			if err != nil{
				return nil, err
			}
			begin = begin / rowLimit
			end = end / rowLimit
			shcnf.RangeExpr = fmt.Sprintf("%d-%d",begin,end)
		}
		//
		mShard, err := NewShard(shcnf, rHash.rangeType,beginVal,rcnf.Format)
		if err != nil {
			return nil, err
		}
		rHash.shards = append(rHash.shards, mShard)
		_,beginVal = mShard.GetRange()
	}
	return rHash,nil
}
func (this *RuleHash)parserRangeExpr(str string) (begin,end int64,err error){
	strArr := strings.Split(str, "-")
	var startNum, endNum int64
	startNum, err = strconv.ParseInt(strArr[0], 10, 64)
	if err != nil {
		return 0,0,err
	}
	if len(strArr) > 1 {
		endNum, err = strconv.ParseInt(strArr[1], 10, 64)
		if err != nil {
			return 0,0,err
		}
	}
	if endNum < startNum {
		t := endNum
		endNum = startNum
		startNum = t
	}
	return startNum,endNum,nil
}
//
func (this *RuleHash)GetShardRule(expr sqlparser.Expr,keyValType string) (rResults []result.RuleResult, err error){
	return this.RuleRange.GetShardRule(expr,keyValType)
}
//======================================================================
func (this *RuleHash)strconvInt64Entity(expr sqlparser.Expr,keyValType string) (val int64,err error){
	if len(keyValType) > 1 && strings.ToLower(keyValType) == "varchar" || strings.ToLower(keyValType) == "string"{
		buf := sqlparser.NewTrackedBuffer(nil)
		expr.Format(buf)
		//use hashcode
		n := crc32.ChecksumIEEE(buf.Bytes())
		buf.Reset()
		val = int64(n) % int64(this.maxLen)
		return
	}
	//
	val,err =  this.RuleRange.strconvInt64Entity(expr,keyValType)
	val = val % int64(this.maxLen)
	//glog.Info("======",val)
	return;
}