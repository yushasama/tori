package monitor

type Variant struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Available bool   `json:"available"`
	Price     string `json:"price"`
	Image     *Image `json:"featured_image,omitempty"`
}

type Image struct {
	Src string `json:"src"`
}

type Product struct {
	Title    string    `json:"title"`
	Variants []Variant `json:"variants"`
}

type ProductsWrapper struct {
	Products []Product `json:"products"`
}
