package officialaccount

import (
	"fmt"
	"github.com/Kee1997/wechat/v2/credential"
	"github.com/Kee1997/wechat/v2/officialaccount"
	offConfig "github.com/Kee1997/wechat/v2/officialaccount/config"
	opContext "github.com/Kee1997/wechat/v2/openplatform/context"
	"github.com/Kee1997/wechat/v2/openplatform/officialaccount/js"
	"github.com/Kee1997/wechat/v2/openplatform/officialaccount/oauth"
	"sync"
)

// OfficialAccount 代公众号实现业务
type OfficialAccount struct {
	// 授权的公众号的appID
	appID string
	*officialaccount.OfficialAccount
}

func (officialAccount *OfficialAccount) SetAppID(appID string) {
	officialAccount.appID = appID
}

// NewOfficialAccount 实例化
// appID :为授权方公众号 APPID，非开放平台第三方平台 APPID
func NewOfficialAccount(opCtx *opContext.Context, appID string) *OfficialAccount {
	officialAccount := officialaccount.NewOfficialAccount(&offConfig.Config{
		AppID:          opCtx.AppID,
		EncodingAESKey: opCtx.EncodingAESKey,
		Token:          opCtx.Token,
		Cache:          opCtx.Cache,
	})
	// 设置获取access_token的函数
	officialAccount.SetAccessTokenHandle(NewDefaultAuthrAccessToken(opCtx, appID))
	return &OfficialAccount{appID: appID, OfficialAccount: officialAccount}
}

// PlatformOauth 平台代发起oauth2网页授权
func (officialAccount *OfficialAccount) PlatformOauth() *oauth.Oauth {
	return oauth.NewOauth(officialAccount.GetContext())
}

// PlatformJs 平台代获取js-sdk配置
func (officialAccount *OfficialAccount) PlatformJs() *js.Js {
	return js.NewJs(officialAccount.GetContext(), officialAccount.appID)
}

// DefaultAuthrAccessToken 默认获取授权ak的方法
type DefaultAuthrAccessToken struct {
	opCtx *opContext.Context
	appID string
	accessTokenLock *sync.Mutex
}

// NewDefaultAuthrAccessToken New
func NewDefaultAuthrAccessToken(opCtx *opContext.Context, appID string) credential.AccessTokenHandle {
	return &DefaultAuthrAccessToken{
		opCtx: opCtx,
		appID: appID,
		accessTokenLock: new(sync.Mutex),
	}
}

// GetAccessToken 获取ak
func (ak *DefaultAuthrAccessToken) GetAccessToken() (string, error) {
	authrTokenKey := fmt.Sprintf("%s_access_token", ak.appID)

	if val := ak.opCtx.Cache.Get(authrTokenKey); val != nil {
		return val.(string), nil
	}

	// 加上lock，是为了防止在并发获取token时，cache刚好失效，导致从微信服务器上获取到不同token
	ak.accessTokenLock.Lock()
	defer ak.accessTokenLock.Unlock()

	// 双检，防止重复从微信服务器获取
	if val := ak.opCtx.Cache.Get(authrTokenKey); val != nil {
		return val.(string), nil
	}
	// cache失效，从微信服务器获取

	refreshToken := fmt.Sprintf("%s_refresh_token", ak.appID)

	refreshTokenValue := ak.opCtx.Cache.Get(refreshToken)

	if refreshTokenValue == nil {
		return "", fmt.Errorf("cannot get authorizer %s refresh token", ak.appID)
	}
	token, err := ak.opCtx.RefreshAuthrToken(ak.appID, refreshTokenValue.(string))
	if err != nil {
		return "", err
	}
	return token.AccessToken, nil
}
