// Wrapper of govmomi

package driver

import (
	"context"
	"fmt"
	"jeffwubj/kubev/pkg/kubev/model"
	"net/url"

	"github.com/vmware/govmomi"
)

func NewClient(ctx context.Context, answers model.Answers) (*govmomi.Client, error) {
	serverurl := url.URL{
		Scheme: "https",
		Path:   fmt.Sprintf("%s:%d/sdk", answers.Serverurl, answers.Port),
		User:   url.UserPassword(answers.Username, answers.Password),
	}
	c, err := govmomi.NewClient(ctx, &serverurl, true)
	if err != nil {
		return nil, err
	}
	return c, nil
}
