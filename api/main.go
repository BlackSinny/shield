package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"
)

type Student struct {
	gorm.Model
	Code  string
	Name  string
	Books []Book `gorm:"many2many:student_books;"`
}

type Book struct {
	gorm.Model
	Name     string
	Students []Student `gorm:"many2many:student_books;"`
}

var db *gorm.DB
var err error

func main() {
	db, err = gorm.Open("mysql", "root:123456.ab@/demo?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Fatal("db connect error")
	}
	defer db.Close()
	db.LogMode(true)
	db.AutoMigrate(&Student{}, &Book{})

	student1 := Student{Code: "000001", Name: "张三"}
	student2 := Student{Code: "000002", Name: "李四"}
	student3 := Student{Code: "000003", Name: "王五"}
	db.Save(&student1)
	db.Save(&student2)
	db.Save(&student3)

	var studentlist []Student
	db.Table("students").Where("id = 1").Or("id = 2").Find(&studentlist)

	book1 := Book{Name: "笑傲江湖", Students: studentlist}
	book2 := Book{
		Name: "神雕侠侣", Students: []Student{
			student3,
		},
	}

	db.Save(&book1)
	db.Save(&book2)

	var student Student
	db.Table("students").Where("id = 1").First(&student)
	book := []Book{}
	db.Preload("Students").Find(&book)
	fmt.Println(book)

	db.Model(&student).Related(&book, "Books")
	fmt.Println(book)
	db.Model(&student).Association("Books").Find(&book)
	fmt.Println(book)

	var bookQ Book
	db.Table("books").Where("id = 1").First(&bookQ)
	db.Model(&bookQ).Association("Students").Find(&studentlist)
	fmt.Println(studentlist)

	db.Model(&bookQ).Association("Students").Append(student3)
	db.Model(&bookQ).Association("Students").Append(Student{Code: "000005", Name: "西门吹雪"})

	db.Model(&bookQ).Association("Students").Delete(student3)
	db.Model(&bookQ).Association("Students").Clear()
	r := gin.Default()
	//r.GET("/client/:client_token", show) //根据id获取用户
	r.GET("/client/:id", show)
	_ = r.Run()
}

func purgeDB(db *gorm.DB) {
	if db.HasTable(&Student{}) {
		db.DropTable(&Student{})
	}
	if db.HasTable(&Book{}) {
		db.DropTable(&Book{})
	}
}

//根据id获取用户
func show(c *gin.Context) {
	var student Student
	id := c.Params.ByName("id")
	db.Debug().Table("students").Where("id = ?", id).First(&student)
	book := []Book{}
	db.Preload("Students").Find(&book)
	fmt.Println(err)
	if student.ID == 0 {
		c.JSON(404, gin.H{"message": "user not found"})
		return
	}
	c.JSON(200, student)
}
