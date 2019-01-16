// Copyright © 2019 Jeff Wu <jeff.wu.junfei@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package driver

import (
	"context"
	"fmt"
	"net/url"

	"github.com/jeffwubj/kubev/pkg/kubev/model"
	"github.com/vmware/govmomi"
)

func NewClient(ctx context.Context, answers *model.Answers) (*govmomi.Client, error) {
	serverurl := url.URL{
		Scheme: "https",
		Path:   "sdk",
		User:   url.UserPassword(answers.Username, answers.Password),
		Host:   fmt.Sprintf("%s:%d", answers.Serverurl, answers.Port),
	}
	c, err := govmomi.NewClient(ctx, &serverurl, true)
	if err != nil {
		return nil, err
	}
	return c, nil
}
