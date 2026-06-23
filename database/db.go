package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

// DB is the global database connection pool.
// يتم استخدام الحرف الكبير (DB) لجعله متاحاً (Exported) لباقي الحزم (Packages) في المشروع.
var DB *sql.DB

// Connect initializes the connection to the PostgreSQL database.
func Connect() {
	var err error

	// 🌟 الممارسة الاحترافية: قراءة بيانات الاتصال من متغيرات البيئة مع قيم افتراضية
	// هذا يمنع تسريب كلمات المرور عند رفع الكود على GitHub
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "123456") // يمكن تغييرها بسهولة من هنا أو من بيئة التشغيل
	dbname := getEnv("DB_NAME", "corehub_db")

	// بناء سلسلة الاتصال بشكل ديناميكي ومنظم
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// فتح الاتصال
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("❌ Failed to open database connection: %v\n", err)
	}

	// التحقق من نجاح الاتصال الفعلي بقاعدة البيانات
	if err = DB.Ping(); err != nil {
		log.Fatalf("❌ Database ping failed: %v\n", err)
	}

	log.Println("✅ Successfully connected to PostgreSQL database!")
}

// getEnv is a helper function to read an environment variable or return a default value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
