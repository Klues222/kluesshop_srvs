package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"mxshop_srvs/user_srv/proto"
)

var userClient proto.UserClient
var conn *grpc.ClientConn

func Init() {
	var err error
	conn, err = grpc.Dial("198.18.0.1:50051", grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	userClient = proto.NewUserClient(conn)
}
func TestGetUserList() {
	rsp, err := userClient.GetUserList(context.Background(), &proto.PageInfo{
		Pn:    1,
		PSize: 2,
	})
	if err != nil {
		panic(err)
	}
	for _, user := range rsp.Data {
		fmt.Println(user.Mobile, user.NickName, user.PassWord)
		checkrsp, err := userClient.CheckPassword(context.Background(), &proto.PasswordCheckInfo{
			PassWord:          "admin123",
			EncryptedPassword: user.PassWord,
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(checkrsp.Success)
	}
}

func TestCreateUser() {
	for i := 0; i < 10; i++ {
		rsp, err := userClient.CreateUser(context.Background(), &proto.CreateUserInfo{
			NickName: fmt.Sprintf("bobby%d", i),
			Mobile:   fmt.Sprintf("1890909899%d", i),
			Password: "admin123",
		})
		if err != nil {
			panic(err)
		}
		fmt.Println(rsp.Id)
	}
}

func TestUpdateUser() {
	_, err := userClient.UpdateUser(context.Background(), &proto.UpdateUserInfo{
		Id:       11,
		NickName: "bobby100",
	})
	if err != nil {
		panic(err)
	}
	println("修改成功")
}
func TestCheckPassword() {
	err, _ := userClient.CheckPassword(context.Background(), &proto.PasswordCheckInfo{
		PassWord:          "admin123",
		EncryptedPassword: "$pbkdf2-sha512$vUVc9ibkgGq5EyXe$d81d2fafebba8722bbc12ef8379475728dacc6e408e70cbf367c9d8be3391250",
	})
	println(err.Success)
}
func TestGetUserMobile() {
	rsp, err := userClient.GetUserMobile(context.Background(), &proto.MobileRequest{Mobile: "18909098991" +
		""})
	if err != nil {
		fmt.Println("用户不存在")

	} else {
		fmt.Println(rsp)
	}
}

func main() {
	Init()
	//TestCreateUser()
	//TestGetUserList()
	//TestUpdateUser()
	//TestCheckPassword()
	TestGetUserMobile()
	conn.Close()
}
