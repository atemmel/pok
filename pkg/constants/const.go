package constants

const (
	TileSize = 16
	DisplaySizeX = 2 * 16 * TileSize
	DisplaySizeY = 2 * 12 * TileSize

	WindowSizeX = DisplaySizeX * 2
	WindowSizeY = DisplaySizeY * 2

	ResourceDir = "resources/"
	TileMapDir =   ResourceDir + "tilemaps/"
	ImagesDir = ResourceDir + "images/"
	FontsDir = ResourceDir + "fonts/"
	AudioDir = ResourceDir + "audio/"
	DialogDir = ResourceDir + "dialog/"
	TileMapImagesDir = ImagesDir + "overworld/"
	CharacterImagesDir = ImagesDir + "characters/"
	PropsImagesDir = ImagesDir + "props/"

	EditorResourceDir = "./editorresources/"
	EditorImagesDir = EditorResourceDir + "images/"
	OverworldObjectsDir = EditorResourceDir + "overworldobjects/"
	AutotileInfoDir = EditorResourceDir + "autotileinfo/"
	TreeAutotileInfoDir = EditorResourceDir + "treeautotileinfo/"

	TopLeftCorner  = 0
	TopRightCorner = 1
	BotLeftCorner  = 2
	BotRightCorner = 3
)
