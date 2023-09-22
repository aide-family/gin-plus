package ginplus

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	oteltrace "go.opentelemetry.io/otel/trace"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type HandleApi struct {
}

func (l *HandleApi) Get() gin.HandlerFunc {
	sqlDB, err := sql.Open("mysql", "root:12345678@tcp(localhost:3060)/prom?charset=utf8mb4&parseTime=True&loc=Local")
	if err != nil {
		panic(err)
	}
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn: sqlDB,
	}))
	if err != nil {
		panic(err)
	}

	db = db.Debug()

	if err := db.Use(&OpentracingPlugin{}); err != nil {
		panic(err)
	}
	return func(c *gin.Context) {
		_span := c.Value("span")
		fmt.Printf("span: %T\n", _span)
		ctx := c.Request.Context()
		span, ok := _span.(oteltrace.Span)
		if ok {
			fmt.Println("--------------ok------------------")
			ctx, span = span.TracerProvider().Tracer("api").Start(ctx, "HandleApi.Get")
			defer span.End()
		}
		type Dict struct {
			Id   uint   `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID"`
			Name string `gorm:"column:name;type:varchar(255);not null;comment:字典名称"`
		}
		list := make([]*Dict, 0)
		err = db.Table("prom_dict").WithContext(ctx).Find(&list).Error
		if err != nil {
			c.String(200, err.Error())
			return
		}

		var first Dict
		if err := db.Table("prom_dict").WithContext(ctx).First(&first).Error; err != nil {
			c.String(200, err.Error())
			return
		}

		if err := db.Table("prom_dict").WithContext(ctx).Where("id = 0").Updates(&first).Error; err != nil {
			c.String(200, err.Error())
			return
		}

		if err := db.Table("prom_dict").WithContext(ctx).Where("id = 0").Delete(&first).Error; err != nil {
			c.String(200, err.Error())
			return
		}

		c.JSON(200, list)
	}
}

func NewHandleApiApi() *HandleApi {
	return &HandleApi{}
}

func TestApiHandleFunc(t *testing.T) {
	ginR := gin.New()

	midd := NewMiddleware(NewResponse())
	serverName := "gin-plus"

	id, _ := os.Hostname()

	ginR.Use(
		midd.Tracing(TracingConfig{
			Name:        serverName,
			URL:         "http://localhost:14268/api/traces",
			Environment: "test",
			ID:          id,
		}),
		midd.Logger(serverName),
		midd.IpLimit(100, 5, "iplimit"),
		midd.Interceptor(),
		midd.Cors(),
	)

	r := New(ginR, WithControllers(NewHandleApiApi()))

	NewCtrlC(r).Start()
}
