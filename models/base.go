package models

func Insert(sec interface{}) error {
	_, err := x.Insert(sec)
	return err
}

func GetById(id interface{}, bean interface{}) (bool, error) {
	return x.Id(id).Get(bean)
}

func GetByExample(bean interface{}) (bool, error) {
	return x.Get(bean)
}

func FindByExample(ptr2slices interface{}, bean interface{}) error {
	return x.Find(ptr2slices, bean)
}

func DelByExample(bean interface{}) error {
	_, err := x.Delete(bean)
	return err
}

func DelById(id, bean interface{}) error {
	_, err := x.Id(id).Delete(bean)
	return err
}

func UpdateById(id interface{}, bean interface{}, cols ...string) error {
	_, err := x.Id(id).Cols(cols...).Update(bean)
	return err
}
