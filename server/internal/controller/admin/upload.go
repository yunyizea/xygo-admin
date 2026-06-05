package admin

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"

	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/google/uuid"

	api "xygo/api/admin"
	"xygo/internal/dao"
	"xygo/internal/library/storager"
	"xygo/internal/library/token"
	"xygo/internal/logic"
	"xygo/internal/model/do"
	"xygo/internal/model/entity"
	"xygo/internal/model/input/adminin"
	"xygo/internal/model/input/form"
	"xygo/utility"
)

// UploadFile 单文件上传
func (c *ControllerV1) UploadFile(ctx context.Context, req *api.UploadFileReq) (res *api.UploadFileRes, err error) {
	r := g.RequestFromCtx(ctx)
	upFile := r.GetUploadFile("file")
	if upFile == nil {
		return nil, gerror.New("未选择文件")
	}
	topic := strings.TrimSpace(req.Topic)

	cfg, err := logic.LoadUploadConfig(ctx)
	if err != nil {
		return nil, err
	}
	drive := strings.ToLower(strings.TrimSpace(req.Drive))
	if drive == "" {
		drive = cfg.Driver
	}

	// 校验大小
	if cfg.MaxSizeBytes > 0 && upFile.Size > cfg.MaxSizeBytes {
		return nil, gerror.Newf("文件大小超过限制：最大 %d 字节", cfg.MaxSizeBytes)
	}
	// 校验后缀
	ext := logic.NormalizeExt(upFile.Filename)
	if len(cfg.AllowedSuffixes) > 0 && !inSlice(ext, cfg.AllowedSuffixes) {
		return nil, gerror.Newf("不允许的文件后缀: %v", ext)
	}
	// 校验 MIME
	if len(cfg.AllowedMimes) > 0 {
		mime := strings.ToLower(strings.TrimSpace(upFile.Header.Get("Content-Type")))
		if mime != "" && !inSlice(mime, cfg.AllowedMimes) {
			return nil, gerror.Newf("不允许的 MIME 类型: %v", mime)
		}
	}

	// 读取内容，计算 sha1
	data, err := readAll(upFile)
	if err != nil {
		return nil, err
	}
	sha1sum := logic.Sha1Bytes(data)
	mimeVal := strings.ToLower(strings.TrimSpace(upFile.Header.Get("Content-Type")))
	if mimeVal == "" {
		mimeVal = http.DetectContentType(data)
	}
	if topic == "" {
		topic = detectTopic(mimeVal, ext)
	}

	width, height := detectImageSize(data)

	// 获取存储驱动
	storage := storager.Instance(ctx)
	drive = storage.DriverName()

	// 先尝试去重复用
	if reused, errReuse := tryReuseAttachment(ctx, cfg.LocalBaseDir, drive, topic, sha1sum); errReuse == nil && reused != nil {
		return reused, nil
	}

	// 使用存储驱动上传
	uploadResult, err := storage.Upload(ctx, &storager.UploadFile{
		Data:     data,
		Filename: upFile.Filename,
		Ext:      ext,
		MimeType: mimeVal,
		Size:     int64(len(data)),
	})
	if err != nil {
		return nil, gerror.Wrapf(err, "上传文件失败")
	}

	res = new(api.UploadFileRes)
	res.UploadFileModel = &adminin.UploadFileModel{
		URL:   uploadResult.FullUrl,
		Path:  uploadResult.RelPath,
		Size:  int64(len(data)),
		Mime:  mimeVal,
		Ext:   "." + ext,
		Drive: drive,
	}

	// 记录附件表
	userId := currentUserID(ctx)
	attachmentId, err := saveAttachmentRecord(ctx, res, upFile.Filename, sha1sum, topic, userId, width, height)
	if err != nil {
		g.Log().Warningf(ctx, "保存附件记录失败: %v", err)
	} else {
		res.AttachmentId = uint64(attachmentId)
	}

	return
}

// AttachmentList 附件列表
func (c *ControllerV1) AttachmentList(ctx context.Context, req *api.AttachmentListReq) (res *api.AttachmentListRes, err error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	size := req.PageSize
	if size <= 0 {
		size = 20
	}

	m := dao.SysAttachment.Ctx(ctx)
	if req.Topic != "" {
		m = m.Where(dao.SysAttachment.Columns().Topic, req.Topic)
	}
	if req.Storage != "" {
		m = m.Where(dao.SysAttachment.Columns().Storage, req.Storage)
	}
	if req.Name != "" {
		m = m.WhereLike(dao.SysAttachment.Columns().Name, "%"+req.Name+"%")
	}

	total, err := m.Count()
	if err != nil {
		return nil, err
	}

	var list []entity.SysAttachment
	if total > 0 {
		if err = m.Page(page, size).OrderDesc(dao.SysAttachment.Columns().Id).Scan(&list); err != nil {
			return nil, err
		}
	}

	items := make([]adminin.AttachmentListItem, 0, len(list))
	for _, it := range list {
		items = append(items, adminin.AttachmentListItem{
			Id:         it.Id,
			Topic:      it.Topic,
			UserId:     it.UserId,
			Url:        it.Url,
			Name:       it.Name,
			Size:       it.Size,
			Mimetype:   it.Mimetype,
			Storage:    it.Storage,
			Sha1:       it.Sha1,
			Quote:      uint(it.Quote),
			Width:      uint(it.Width),
			Height:     uint(it.Height),
			CreateTime: uint(it.CreateTime),
			UpdateTime: uint(it.UpdateTime),
		})
	}

	res = new(api.AttachmentListRes)
	res.AttachmentListModel = &adminin.AttachmentListModel{
		List: items,
		PageRes: form.PageRes{
			Page:     page,
			PageSize: size,
			Total:    total,
		},
	}
	return
}

func readAll(f *ghttp.UploadFile) ([]byte, error) {
	reader, err := f.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(reader)
}

// detectTopic 根据 mime/后缀分类
func detectTopic(mime, ext string) string {
	ext = strings.ToLower(ext)
	mime = strings.ToLower(mime)
	if mime == "" && ext == "" {
		return "other"
	}
	if strings.HasPrefix(mime, "image/") || inSlice(ext, []string{"jpg", "jpeg", "png", "gif", "webp", "bmp", "svg"}) {
		return "image"
	}
	if strings.HasPrefix(mime, "audio/") || inSlice(ext, []string{"mp3", "wav", "wma", "aac", "ogg", "flac"}) {
		return "audio"
	}
	if strings.HasPrefix(mime, "video/") || inSlice(ext, []string{"mp4", "avi", "mov", "wmv", "flv", "mkv", "webm"}) {
		return "video"
	}
	if inSlice(ext, []string{"zip", "rar", "7z", "tar", "gz"}) {
		return "archive"
	}
	if inSlice(ext, []string{"doc", "docx", "xls", "xlsx", "ppt", "pptx", "pdf", "txt", "md"}) {
		return "doc"
	}
	return "other"
}

// detectImageSize 读取图片宽高
func detectImageSize(data []byte) (w, h int) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0
	}
	return cfg.Width, cfg.Height
}

// currentUserID 从 token 获取用户ID（后台管理员）
func currentUserID(ctx context.Context) uint {
	r := g.RequestFromCtx(ctx)
	if r == nil {
		return 0
	}
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return 0
	}
	tokenStr := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer"))
	if tokenStr == "" {
		return 0
	}
	authUser, err := token.Parse(ctx, tokenStr)
	if err != nil || authUser == nil {
		return 0
	}
	return uint(authUser.Id)
}

// saveLocalFromBytes 将上传文件保存到本地资源目录
func saveLocalFromBytes(base string, data []byte, ext string) (relPath string, fullPath string, err error) {
	if ext == "" {
		ext = "bin"
	}
	subdir := gtime.Now().Format("Ymd")
	// 生成相对路径：固定前缀 attachment/upload/yyyyMMdd/uuid.ext
	name := uuid.New().String() + "." + ext
	relPath = gfile.Join("attachment", "upload", subdir, name)
	relPath = strings.ReplaceAll(relPath, `\`, `/`)

	// 生成物理路径：base + subdir + name
	fullPath = gfile.Join(base, subdir, name)
	if err = gfile.Mkdir(gfile.Dir(fullPath)); err != nil {
		return
	}

	if err = gfile.PutBytes(fullPath, data); err != nil {
		return
	}
	return
}

// pathFromBase 生成相对 URL 路径：去掉 base 前缀，并统一 / 分隔
func pathFromBase(base, full string) string {
	b := gfile.RealPath(base)
	if b == "" {
		b = base
	}
	f := gfile.RealPath(full)
	if f == "" {
		f = full
	}
	p := strings.TrimPrefix(f, b)
	p = strings.TrimLeft(p, `\/`)
	return strings.ReplaceAll(p, `\`, `/`)
}

// saveAttachmentRecord 写入附件记录
func saveAttachmentRecord(ctx context.Context, res *api.UploadFileRes, originalName string, sha1sum string, topic string, userId uint, width, height int) (int64, error) {
	// PostgreSQL 驱动不支持 LastInsertId()，这里先 Insert，再通过 sha1+storage+topic 回查 ID
	_, err := dao.SysAttachment.Ctx(ctx).Data(do.SysAttachment{
		Topic:      topic,
		UserId:     uint64(userId),
		Url:        "/" + strings.TrimLeft(res.Path, "/"),
		Width:      uint(width),
		Height:     uint(height),
		Name:       originalName,
		Size:       uint64(res.Size),
		Mimetype:   res.Mime,
		Quote:      1,
		Storage:    res.Drive,
		Sha1:       sha1sum,
		CreateTime: uint(utility.NowUnix()),
		UpdateTime: uint(utility.NowUnix()),
	}).Insert()

	if err != nil {
		return 0, err
	}

	// 回查新插入的附件记录 ID（兼容 PostgreSQL）
	var record entity.SysAttachment
	err = dao.SysAttachment.Ctx(ctx).
		Where(dao.SysAttachment.Columns().Sha1, sha1sum).
		Where(dao.SysAttachment.Columns().Storage, res.Drive).
		Where(dao.SysAttachment.Columns().Topic, topic).
		OrderDesc(dao.SysAttachment.Columns().Id).
		Scan(&record)
	if err != nil {
		return 0, err
	}
	return int64(record.Id), nil
}

// tryReuseAttachment 通过 sha1 + storage + topic 查找附件，存在且文件存在则复用
func tryReuseAttachment(ctx context.Context, base, storage, topic, sha1sum string) (*api.UploadFileRes, error) {
	rec := &do.SysAttachment{}
	err := dao.SysAttachment.Ctx(ctx).
		Where(dao.SysAttachment.Columns().Sha1, sha1sum).
		Where(dao.SysAttachment.Columns().Storage, storage).
		Where(dao.SysAttachment.Columns().Topic, topic).
		Scan(&rec)
	if err != nil {
		if strings.Contains(err.Error(), "doesn't exist") {
			return nil, nil
		}
		return nil, err
	}
	if rec == nil {
		return nil, nil
	}
	urlVal := strings.TrimLeft(gconv.String(rec.Url), "/")
	full := gfile.Join(base, urlVal)
	if !gfile.Exists(full) {
		return nil, nil
	}
	// 增加引用次数
	_, _ = dao.SysAttachment.Ctx(ctx).
		Where(dao.SysAttachment.Columns().Id, rec.Id).
		Data(g.Map{
			dao.SysAttachment.Columns().Quote:      gconv.Int(rec.Quote) + 1,
			dao.SysAttachment.Columns().UpdateTime: utility.UnixToGTime(utility.NowUnix()),
		}).Update()
	reusedRes := new(api.UploadFileRes)
	reusedRes.UploadFileModel = &adminin.UploadFileModel{
		URL:   "/" + urlVal,
		Path:  urlVal,
		Size:  gconv.Int64(rec.Size),
		Mime:  gconv.String(rec.Mimetype),
		Ext:   "." + logic.NormalizeExt(gconv.String(rec.Name)),
		Drive: storage,
	}
	return reusedRes, nil
}

// AttachmentDelete 删除附件
func (c *ControllerV1) AttachmentDelete(ctx context.Context, req *api.AttachmentDeleteReq) (res *api.AttachmentDeleteRes, err error) {
	// 查询附件信息
	var attachment *entity.SysAttachment
	err = dao.SysAttachment.Ctx(ctx).Where("id", req.Id).Scan(&attachment)
	if err != nil {
		return nil, err
	}
	if attachment == nil {
		return nil, gerror.New("附件不存在")
	}

	// 删除数据库记录
	_, err = dao.SysAttachment.Ctx(ctx).Where("id", req.Id).Delete()
	if err != nil {
		return nil, err
	}

	// 删除物理文件（当引用数 <= 1 时）
	if attachment.Quote <= 1 && attachment.Url != "" {
		storage := storager.Instance(ctx)
		if delErr := storage.Delete(ctx, attachment.Url); delErr != nil {
			g.Log().Warningf(ctx, "[Attachment] delete physical file error: %v, path: %s", delErr, attachment.Url)
		} else {
			g.Log().Infof(ctx, "[Attachment] deleted physical file: %s", attachment.Url)
		}
	}

	return &api.AttachmentDeleteRes{}, nil
}

func inSlice(v string, arr []string) bool {
	for _, a := range arr {
		if v == a {
			return true
		}
	}
	return false
}
