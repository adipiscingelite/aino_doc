package routes

import (
	"document/controller"
	"document/middleware"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Error struct {
	Code    int
	Message string
}

type Handler func(http.ResponseWriter, *http.Request) *Error

func (fn Handler) ServeHTTP(c echo.Context) error {
	w := c.Response().Writer
	r := c.Request()
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if e := fn(w, r); e != nil {
		return c.String(e.Code, e.Message)
	}
	return nil
}
func Route() *echo.Echo {
	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Access-Control-Allow-Origin", "*")
			c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
			c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			return next(c)
		}
	})

	// protectedAssets := e.Group("/assets")
	// protectedAssets.Use(middleware.AuthMiddleware)

	e.Static("/assets", "assets")
	superAdmin := e.Group("/superadmin")
	superAdmin.Use(middleware.SuperAdminMiddleware)

	adminMember := e.Group("/api")
	adminMember.Use(middleware.AdminMemberMiddleware)

	//admin
	adminGroup := e.Group("/admin")
	adminGroup.Use(middleware.AdminMiddleware)
	adminGroup.GET("/my/form/division", controller.FormByDivision)
	adminGroup.GET("/my/itcm/division", controller.FormITCMByDivision)
	adminGroup.GET("/my/da/division", controller.FormDAByDivision)
	adminGroup.GET("/my/ba/division", controller.FormBAByDivision)
	adminGroup.GET("/my/ha/req/division", controller.FormHAByDivision)
	adminGroup.GET("/my/ha/review/division", controller.FormHAByDivisionReview)

	//document
	e.GET("/document", controller.GetAllDoc)
	e.GET("/document/:id", controller.ShowDocById)
	superAdmin.POST("/document/add", controller.AddDocument)
	superAdmin.PUT("/document/update/:id", controller.UpdateDocument)
	superAdmin.PUT("/document/delete/:id", controller.DeleteDoc)

	//semua formulir, tanpa ada by ITCM, DA, BA. campur
	e.GET("/form", controller.GetAllForm)
	e.GET("/form/:id", controller.ShowFormById)
	adminMember.POST("/form/add", controller.AddForm)
	adminMember.GET("/my/form", controller.MyForm)
	adminMember.PUT("/form/update/:id", controller.UpdateForm)

	//tandatangan
	e.GET("/signatory/:id", controller.GetSpecSignatureByID)
	// buat true false
	adminMember.PUT("/signature/update/:id", controller.UpdateSignature)
	e.GET("/form/signatories/:id", controller.GetSignatureForm)
	//add informasi signature
	adminMember.POST("/add/sign/info", controller.AddSignInfo)
	//update informasi signature
	adminMember.PUT("/sign/info/update/:id", controller.UpdateSignInfo)
	//delete info sign
	adminMember.PUT("/sign/info/delete/:id", controller.DeleteSignInfo)
	//list dokumen yang harus ditandatangan
	//da
	adminMember.GET("/my/signature/da", controller.SignatureUser)
	//ba
	adminMember.GET("/my/signature/ba", controller.SignatureUserBA)
	//itcm
	adminMember.GET("/my/signature/itcm", controller.SignatureUserITCM)
	//ha
	adminMember.GET("/my/signature/ha", controller.SignatureUserHA)

	//add approval
	adminMember.PUT("/form/approval/:id", controller.AddApproval)
	//add approval
	adminMember.PUT("/form/da/approval/:id", controller.AddApprovalDA)

	//FORM itcm
	adminMember.POST("/add/itcm", controller.AddITCM)
	e.GET("/form/itcm/code", controller.GetITCMCode)
	e.GET("/form/itcm", controller.GetAllFormITCM)
	e.GET("/form/itcm/:id", controller.GetSpecITCM)
	e.GET("/itcm/:id", controller.GetSpecAllITCM)
	adminMember.PUT("/form/itcm/update/:id", controller.UpdateFormITCM)
	adminMember.GET("/my/form/itcm", controller.GetAllFormITCMbyUserID)
	adminGroup.GET("/itcm/all", controller.GetAllFormITCMAdmin)

	//form BA
	adminMember.POST("/add/ba", controller.AddBA)
	e.GET("/form/ba/code", controller.GetBACode)
	e.GET("/form/ba", controller.GetAllFormBA)
	e.GET("/form/ba/:id", controller.GetSpecBA)
	e.GET("/ba/:id", controller.GetSpecAllBA)
	adminMember.GET("/my/form/ba", controller.GetAllFormBAbyUserID)
	adminGroup.GET("/ba/all", controller.GetAllFormBAAdmin)
	adminMember.PUT("/form/ba/update/:id", controller.UpdateFormBA)

	// ba asset
	adminMember.POST("/add/ba/asset", controller.AddBAAsset)
	e.GET("/form/beritaacara", controller.GetAllFormBAAssets)
	e.GET("/form/beritaacara/:id", controller.GetSpecAllBAAssets)
	adminMember.PUT("/form/beritaacara/update/:id", controller.UpdateBeritaAcara)
	adminMember.PUT("/beritaacara/delete/:id", controller.DeleteBeritaAcara)


	// asset
	adminMember.POST("/add/asset", controller.AddAsset)
	e.GET("/assets", controller.GetAllAssets)
	e.GET("/asset/:id", controller.GetSpecAllAsset)
	adminMember.PUT("/asset/update/:id", controller.UpdateAsset)
	adminMember.PUT("/asset/delete/:id", controller.DeleteAsset)

	//form DA
	adminMember.POST("/add/da", controller.AddDA)
	e.GET("/form/da/code", controller.GetDACode)
	e.GET("/dampak/analisa", controller.GetAllFormDA)
	e.GET("/dampak/analisa/:id", controller.GetSpecDA)
	e.GET("/da/:id", controller.GetSpecAllDAa)
	e.GET("/spec/da/:id", controller.GetSpecAllDA)
	adminMember.PUT("/dampak/analisa/update/:id", controller.UpdateFormDA)
	adminMember.GET("/my/form/da", controller.GetAllFormDAbyUser)
	adminGroup.GET("/da/all", controller.GetAllDAbyAdmin)

	//form hak akses
	e.GET("/form/ha/code", controller.GetHACode)

	//form hak akses permintaan/penghapusan
	adminMember.POST("/add/ha", controller.AddHA)
	e.GET("/hak/akses", controller.GetAllFormHA)
	e.GET("/ha/:id", controller.GetSpecAllHA)
	adminMember.PUT("/hak/akses/update/:id", controller.UpdateHakAkses)
	adminMember.GET("/my/form/ha", controller.MyFormsHA)
	// adminGroup.GET("/ha/all", controller.GetAllFormHAAdmin)
	// adminMember.PUT("/ha/publish/:id", controller.PublishHA)

	//form hak akses review
	adminMember.POST("/add/ha/review", controller.AddHAReview)
	e.GET("/hak/akses/review", controller.GetAllFormHAReview)
	e.GET("/ha/review/:id", controller.GetSpecAllHAReview)
	adminMember.PUT("/hak/akses/review/update/:id", controller.UpdateHakAksesReview)
	adminMember.GET("/my/form/ha/review", controller.MyFormsHAReview)
	adminGroup.GET("/ha/all", controller.GetAllFormHAReviewAdmin)
	// adminMember.PUT("/ha/publish/:id", controller.PublishHA)

	//product
	e.GET("/product", controller.GetAllProduct)
	e.GET("/product/:id", controller.ShowProductById)
	superAdmin.POST("/product/add", controller.AddProduct)
	superAdmin.PUT("/product/update/:id", controller.UpdateProdcut)
	superAdmin.PUT("/product/delete/:id", controller.DeleteProduct)

	//project
	e.GET("/project", controller.GetAllProject)
	e.GET("/project/:id", controller.ShowProjectById)
	superAdmin.POST("/project/add", controller.AddProject)
	superAdmin.PUT("/project/update/:id", controller.UpdateProject)
	superAdmin.PUT("/project/delete/:id", controller.DeleteProject)

	// notif
	adminMember.GET("/my/notif", controller.SignatureNotif)
	adminMember.GET("/my/approve/notif", controller.ApproveNotif)

	//delete form (bisa digunakan untuk semua formulir da, ba, itcm)
	adminMember.PUT("/form/delete/:id", controller.DeleteForm)

	//detail. ga kepake
	e.GET("/detail/itcm/:id", controller.DetailITCM)

	// adminMember.PUT("/upload", controller.woilahuploadSignature)

	return e
}
