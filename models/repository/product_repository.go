package repository

import (
	. "github.com/o0khoiclub0o/piflab-store-api-go/lib"
	. "github.com/o0khoiclub0o/piflab-store-api-go/models"
	. "github.com/o0khoiclub0o/piflab-store-api-go/services"

	"fmt"
	"strings"
	"time"
)

type ProductRepository struct {
	*DB
}

func (repo ProductRepository) FindByListId(ids []uint64) *ProductListId {
	products := &ProductSlice{}
	repo.DB.Order("id asc").Where(ids).Find(products)

	var id_has []uint64
	for idx, product := range *products {
		(*products)[idx].GetImageUrl()
		id_has = append(id_has, uint64(product.Id))
	}

	// detect the product not returned from repo
	var id_not_have []uint64
	for _, val := range ids {
		for _, val_has := range id_has {
			if val_has == val {
				// if the value is match, continue to next value
				goto pass_next
			}
		}
		// the id is not found if it goes there
		id_not_have = append(id_not_have, val)
	pass_next:
		continue
	}

	var err string
	if len(id_not_have) != 0 {
		err += strings.Trim(strings.Replace(fmt.Sprint(id_not_have), " ", ",", -1), "[]")
	}

	return &ProductListId{
		ProductSlice: *products,
		Error:        err,
	}
}

func (repo ProductRepository) FindById(id uint) (*Product, error) {
	product := &Product{}

	err := repo.DB.First(&product, id).Error
	if err != nil {
		return nil, err
	}

	product.GetImageUrl()

	return product, err
}

func (repo ProductRepository) GetAll() (*ProductSlice, error) {
	products := &ProductSlice{}
	err := repo.DB.Find(products).Error

	for idx := range *products {
		(*products)[idx].GetImageUrl()
	}

	return products, err
}

func (repo ProductRepository) GetPage(offset uint, limit uint, sort_field string, sort_order string, search string) (*ProductSlice, uint, error) {
	products := &ProductSlice{}
	var err error
	var lower_search string

	if search == "" {
		err = repo.DB.Order(sort_field + " " + sort_order).Offset(int(offset)).Limit(int(limit)).Find(products).Error
	} else {
		lower_search = strings.ToLower(search)
		err = repo.DB.Order(sort_field + " " + sort_order).Offset(int(offset)).Limit(int(limit)).Where("LOWER(name) LIKE '%" + lower_search + "%'").Find(products).Error
	}

	for idx := range *products {
		(*products)[idx].GetImageUrl()
	}

	count := uint(0)
	if search == "" {
		repo.DB.Table("products").Order(sort_field + " " + sort_order).Count(&count)
	} else {
		repo.DB.Table("products").Order(sort_field + " " + sort_order).Where("LOWER(name) LIKE '%" + lower_search + "%'").Count(&count)
	}

	return products, count, err
}

func (repo ProductRepository) saveFile(product *Product) error {
	type image_to_save struct {
		data  *[]byte
		field ImageField
		size  ImageSize
	}

	if product.Image != "" {
		var images = []image_to_save{
			{&product.ImageData, IMAGE, ORIGIN},
			{&product.ImageThumbnailData, IMAGE, THUMBNAIL},
			{&product.ImageDetailData, IMAGE, DETAIL}}

		for _, image := range images {
			if err := (FileService{}).SaveFile(
				*image.data,
				product.GetImagePath(image.field, image.size),
				product.GetImageContentType(image.field, image.size)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (repo ProductRepository) deleteFile(product *Product) error {
	var fields = []ImageField{IMAGE}
	var sizes = []ImageSize{ORIGIN, THUMBNAIL, DETAIL}

	for _, field := range fields {
		for _, size := range sizes {
			if err := (FileService{}).DeleteFile(product.GetImagePath(field, size)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (repo ProductRepository) createProduct(product *Product) error {
	product.ImageUpdatedAt = time.Now()

	tx := repo.DB.Begin()

	if err := tx.Create(product).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := repo.saveFile(product); err != nil {
		tx.Rollback()
		return err
	}
	product.GetImageUrl()

	tx.Commit()

	return nil
}

func (repo ProductRepository) updateProduct(product *Product) error {
	tx := repo.DB.Begin()

	if product.ImageData != nil {
		if err := repo.deleteFile(product); err != nil {
			tx.Rollback()
			return err
		}

		product.Image = product.NewImage
		product.ImageUpdatedAt = time.Now()

		if err := repo.saveFile(product); err != nil {
			tx.Rollback()
			return err
		}
		product.GetImageUrl()
	}

	if err := tx.Save(product).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()

	return nil
}

func (repo ProductRepository) SaveProduct(product *Product) error {
	if product.Id == 0 {
		return repo.createProduct(product)
	}
	return repo.updateProduct(product)
}

func (repo ProductRepository) CountProduct() (uint, error) {
	count := uint(0)

	err := repo.DB.Table("products").Count(&count).Error

	return count, err
}

func (repo ProductRepository) DeleteProduct(id uint64) (*Product, error) {
	product, err := repo.FindById(uint(id))
	if err != nil {
		return product, err
	}

	tx := repo.DB.Begin()

	if err := repo.deleteFile(product); err != nil {
		tx.Rollback()
		return product, err
	}

	if err := repo.DB.Delete(product).Error; err != nil {
		tx.Rollback()
		return product, err
	}

	tx.Commit()

	return product, nil
}
