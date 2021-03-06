package entity

import "github.com/jinzhu/gorm"

// UpdatePhotoCounts updates photos count in related tables as needed.
func UpdatePhotoCounts() error {
	log.Info("index: updating photo counts")

	if err := Db().Table("places").
		UpdateColumn("photo_count", gorm.Expr("(SELECT COUNT(*) FROM photos ph "+
			"WHERE places.place_uid = ph.place_uid "+
			"AND ph.photo_quality >= 0 "+
			"AND ph.photo_private = 0 "+
			"AND ph.deleted_at IS NULL)")).Error; err != nil {
		return err
	}

	/* See internal/entity/views.go

	CREATE OR REPLACE VIEW label_counts AS
	SELECT label_id, SUM(photo_count) AS photo_count FROM (
	(SELECT l.id AS label_id, COUNT(*) AS photo_count FROM labels l
	            JOIN photos_labels pl ON pl.label_id = l.id
	            JOIN photos ph ON pl.photo_id = ph.id
				WHERE pl.uncertainty < 100
				AND ph.photo_quality >= 0
				AND ph.photo_private = 0
				AND ph.deleted_at IS NULL GROUP BY l.id)
	UNION ALL
	(SELECT l.id AS label_id, COUNT(*) AS photo_count FROM labels l
	            JOIN categories c ON c.category_id = l.id
	            JOIN photos_labels pl ON pl.label_id = c.label_id
	            JOIN photos ph ON pl.photo_id = ph.id
				WHERE pl.uncertainty < 100
				AND ph.photo_quality >= 0
				AND ph.photo_private = 0
				AND ph.deleted_at IS NULL GROUP BY l.id)) counts GROUP BY label_id
	*/

	/* TODO: Requires proper view support in TiDB

	if err := Db().
		Table("labels").
		UpdateColumn("photo_count",
			gorm.Expr("(SELECT photo_count FROM label_counts WHERE label_id = labels.id)")).Error; err != nil {
		log.Warn(err)
	} */

	if err := Db().
		Table("labels").
		UpdateColumn("photo_count",
			gorm.Expr(`(SELECT photo_count FROM (
			SELECT label_id, SUM(photo_count) AS photo_count FROM (
			(SELECT l.id AS label_id, COUNT(*) AS photo_count FROM labels l
			            JOIN photos_labels pl ON pl.label_id = l.id
			            JOIN photos ph ON pl.photo_id = ph.id
						WHERE pl.uncertainty < 100
						AND ph.photo_quality >= 0
						AND ph.photo_private = 0
						AND ph.deleted_at IS NULL GROUP BY l.id)
			UNION ALL
			(SELECT l.id AS label_id, COUNT(*) AS photo_count FROM labels l
			            JOIN categories c ON c.category_id = l.id
			            JOIN photos_labels pl ON pl.label_id = c.label_id
			            JOIN photos ph ON pl.photo_id = ph.id
						WHERE pl.uncertainty < 100
						AND ph.photo_quality >= 0
						AND ph.photo_private = 0
						AND ph.deleted_at IS NULL GROUP BY l.id)) counts GROUP BY label_id
			) label_counts WHERE label_id = labels.id)`)).Error; err != nil {
		return err
	}

	return nil
}
