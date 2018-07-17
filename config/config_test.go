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

package config

import (
	"fmt"
	"testing"
)

func TestXmlConfig(t *testing.T){

	mConfig,err := ParseConfig("conf.xml",true)
	if err != nil{
		fmt.Println(err)
		return
	}
	fmt.Println(mConfig)
	//reg,err  := regexp.Compile("(^\\.\\/|^[a-zA-Z_][^\\:\\/])+")
	//if err != nil{
	//	fmt.Println(err)
	//	return
	//}
	////D:\workspace\golang\tt.log
	//if reg.MatchString("d:/conf/dealer_info.sql"){
	//	fmt.Println("Ok",runtime.GOOS)
	//}
}


