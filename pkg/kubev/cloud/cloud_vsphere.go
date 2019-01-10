// Wrapper of govmomi

package cloud

import (
	"context"
	"jeffwubj/kubev/pkg/kubev/model"

	"github.com/vmware/govmomi"
)

func NewClient(ctx context.Context, answers model.Answers) (*govmomi.Client, error) {

	// work in progress

	// fmt.Println("create new client: ", session)
	// fixedUrl := "https://administrator%40vsphere.local:VMw%40re.c0m@10.117.172.110/sdk"

	// username := answers.Username
	// password := answers.Password

	// username = url.PathEscape(username)
	// password = url.PathEscape(password)

	// serverurl := fmt.Sprintf("https://%s:%s@%s:%d/sdk", username, password, answers.Port)

	// u, _ := url.Parse(serverurl)

	// c, err := govmomi.NewClient(ctx, u, true)

	// // fmt.Println("end NewClient")

	// // Connect and log in to ESX or vCenter
	// return c, err
}
