package context

import (
	"github.com/Kee1997/wechat/v2/credential"
	"github.com/Kee1997/wechat/v2/work/config"
)

// Context struct
type Context struct {
	*config.Config
	credential.AccessTokenHandle
}
