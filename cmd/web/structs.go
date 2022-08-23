package web

type ViewData struct {
	Auth        string
	User        User
	Articles    []Article
	ErrorType   string
	Profile     User
	OneArticle  Article
	Comments    []Comment
	CntArticles int
	CntComments int
	PageId      int
	PageIdNext  int
	PageIdBack  int
}

type User struct {
	Id        int    `json:"id"`
	Nick      string `json:"nick" valid:"required"`
	Name      string `json:"name" valid:"required"`
	Surname   string `json:"surname"`
	Email     string `json:"email" valid:"required"`
	Password  string `json:"password" valid:"required"`
	PhotoPath string `json:"photo"`
	IsAdmin   bool   `json:"is_admin" valid:"required"`
}

type Article struct {
	Id       int    `json:"id"`
	Title    string `json:"title"`
	Text     string `json:"text"`
	Picture  string `json:"picture"`
	AuthorId int    `json:"author_id"`
	Date     string `json:"date"`
	CntLikes int    `json:"cntlikes"`
	Liked    bool   `json:"liked"`
	IsAuthor bool   `json:"is_author"`
}

type Comment struct {
	Id              int    `json:"id"`
	Text            string `json:"text"`
	ArticleId       int    `json:"article_id"`
	AuthorId        int    `json:"author_id"`
	Date            string `json:"date"`
	AuthorName      string `json:"author_name"`
	AuthorSurname   string `json:"author_surname"`
	AuthorNick      string `json:"author_nick"`
	AuthorPhotoPath string `json:"author_photo_path"`
	IsAuthor        bool   `json:"is_author"`
}
