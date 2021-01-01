package store

type Object struct {
	Key   string   `bson:"key"`
	Words []string `bson:"words"`
}
