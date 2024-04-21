package fat_secret

type Response struct {
	Food_Entries FoodEntries `json:"food_entries`
}

type FoodEntries struct {
	Food_Entry []FoodEntry `json:"food_entry"`
}

type FoodEntry struct {
	calcium                string
	Calories               string `json:"calories"`
	carbohydrate           string
	cholesterol            string
	date_int               string
	fat                    string
	fiber                  string
	food_entry_description string
	food_entry_id          string
	food_entry_name        string
	food_id                string
	iron                   string
	meal                   string
	monounsaturated_fat    string
	number_of_units        string
	polyunsaturated_fat    string
	potassium              string
	protein                string
	saturated_fat          string
	serving_id             string
	sodium                 string
	sugar                  string
	vitamin_a              string
	vitamin_c              string
}
