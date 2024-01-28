package grpc_proxy_middleware

import (
	"log"
	"strings"
	"github.com/hugokung/micro_gateway/internal/dao"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"github.com/hugokung/micro_gateway/pkg/public"
)

func GrpcJwtAuthTokenMiddleware(serviceDetail *dao.ServiceDetail) func(interface{}, grpc.ServerStream, *grpc.StreamServerInfo, grpc.StreamHandler) error  {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

		//fmt.Println("serviceDetail",serviceDetail)
		// decode jwt token
		// app_id 与  app_list 取得 appInfo
		// appInfo 放到 gin.context
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return errors.New("miss metadata from context")
		}
		authToken := ""
		auths := md.Get("authorization")
		if len(auths) > 0 {
			authToken = auths[0]
		}
		token:=strings.ReplaceAll(authToken, "Bearer ", "")
		//fmt.Println("token",token)
		appMatched := false
		if token != ""{
			claims, err:=public.JwtDecode(token)
			if err != nil{
				return err
			}
			//fmt.Println("claims.Issuer",claims.Issuer)
			appList:=dao.AppManagerHandler.GetAppList()
			for _,appInfo:=range appList{
				if appInfo.AppID == claims.Issuer{
					md.Set("app", public.Obj2Json(appInfo))
					appMatched = true
					break
				}
			}
		}
		if serviceDetail.AccessControl.OpenAuth==1 && !appMatched{
			return errors.New("not match valid app")
		}
		if err := handler(srv, ss); err != nil {
			log.Printf("RPC failed with error %v\n", err)
			return err
		}
		return nil
	}
}
