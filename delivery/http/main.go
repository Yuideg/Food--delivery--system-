package main

import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/yuidegm/web-prog-go-sample-master/delivery/http/handler"
	"github.com/yuidegm/web-prog-go-sample-master/entity"
	"github.com/yuidegm/web-prog-go-sample-master/rtoken"
)

func createTables(dbconn *gorm.DB) []error {
	errs := dbconn.CreateTable(&entity.User{}, &entity.Role{}, &entity.Session{}, &entity.Item{}, &entity.Order{}, &entity.Category{}, &entity.Ingredient{}, &entity.Comment{}).GetErrors()
	if errs != nil {
		return errs
	}
	return nil
}

func main() {
	//createTables(dbconn)

	csrfSignKey := []byte(rtoken.GenerateRandomID(32))
	tmpl := template.Must(template.ParseGlob("ui/templates/*"))

	connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", os.Getenv("DBUSER"),
		os.Getenv("DBPASS"), os.Getenv("DBHOST"), os.Getenv("DBNAME"))
	dbconn, err := gorm.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	defer dbconn.Close()

	sessionRepo := urepimp.NewSessionGormRepo(dbconn)
	sessionSrv := usrvimp.NewSessionService(sessionRepo)

	categoryRepo := mrepim.NewCategoryGormRepo(dbconn)
	categoryServ := msrvim.NewCategoryService(categoryRepo)

	userRepo := urepimp.NewUserGormRepo(dbconn)
	userServ := usrvimp.NewUserService(userRepo)

	roleRepo := urepimp.NewRoleGormRepo(dbconn)
	roleServ := usrvimp.NewRoleService(roleRepo)

	ach := handler.NewAdminCategoryHandler(tmpl, categoryServ, csrfSignKey)
	mh := handler.NewMenuHandler(tmpl, categoryServ, csrfSignKey)

	sess := configSess()
	uh := handler.NewUserHandler(tmpl, userServ, sessionSrv, roleServ, sess, csrfSignKey)

	fs := http.FileServer(http.Dir("ui/assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	http.HandleFunc("/", mh.Index)
	http.HandleFunc("/about", mh.About)
	http.HandleFunc("/contact", mh.Contact)
	http.HandleFunc("/menu", mh.Menu)
	http.Handle("/admin", uh.Authenticated(uh.Authorized(http.HandlerFunc(mh.Admin))))

	http.Handle("/admin/categories", uh.Authenticated(uh.Authorized(http.HandlerFunc(ach.AdminCategories))))
	http.Handle("/admin/categories/new", uh.Authenticated(uh.Authorized(http.HandlerFunc(ach.AdminCategoriesNew))))
	http.Handle("/admin/categories/update", uh.Authenticated(uh.Authorized(http.HandlerFunc(ach.AdminCategoriesUpdate))))
	http.Handle("/admin/categories/delete", uh.Authenticated(uh.Authorized(http.HandlerFunc(ach.AdminCategoriesDelete))))

	http.Handle("/admin/users", uh.Authenticated(uh.Authorized(http.HandlerFunc(uh.AdminUsers))))
	http.Handle("/admin/users/new", uh.Authenticated(uh.Authorized(http.HandlerFunc(uh.AdminUsersNew))))
	http.Handle("/admin/users/update", uh.Authenticated(uh.Authorized(http.HandlerFunc(uh.AdminUsersUpdate))))
	http.Handle("/admin/users/delete", uh.Authenticated(uh.Authorized(http.HandlerFunc(uh.AdminUsersDelete))))

	http.HandleFunc("/login", uh.Login)
	http.Handle("/logout", uh.Authenticated(http.HandlerFunc(uh.Logout)))
	http.HandleFunc("/signup", uh.Signup)

	port := fmt.Sprintf(":%s", os.Getenv("HPORT"))

	http.ListenAndServe(port, nil)
}

func configSess() *entity.Session {
	tokenExpires := time.Now().Add(time.Minute * 30).Unix()
	sessionID := rtoken.GenerateRandomID(32)
	signingString, err := rtoken.GenerateRandomString(32)
	if err != nil {
		panic(err)
	}
	signingKey := []byte(signingString)

	return &entity.Session{
		Expires:    tokenExpires,
		SigningKey: signingKey,
		UUID:       sessionID,
	}
}
