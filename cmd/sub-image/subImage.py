from PIL import Image, ImageDraw

class BoundingBox:
    
    def __init__(self, topP, botP) -> None:
        self.topy, self.topx = topP
        self.boty, self.botx = botP
        print("new box")

    def getRect(self):
        return [(self.topx, self.topy), (self.botx, self.boty)]


    def expand_box(self, point):
        #yx
        if point[1] < self.topx:
            self.topx = point[1]        
        if point[1] > self.botx:
            self.botx = point[1]
        if point[0] < self.topy:
            self.topy = point[0]
        if point[0] > self.boty:
            self.boty = point[0]

    def in_box(self, point):
        if (point[1] >= self.topx and point[1] <= self.botx) and (point[0] >= self.topy and point[0] <= self.boty):
            return True
        else:
            return False

def find_in_list(elem, list):
    for i, pixel in enumerate(list):
        if elem == pixel[0]:
            return i
    return -1

def neighbour(x, y, p_index, size, coordinates):
    new_point = [y, x]
    new_neighbour = coordinates[find_in_list(new_point, coordinates)]
    if new_neighbour[1] != (0,0,0,0) and (new_point[p_index] <= size and new_point[p_index] >= 0):
        return new_neighbour

def get_neighbours(point, coordinates, img_size):
    neighbours = []
    y, x = point[0], point[1]
    #only add neighbours who hasnt been marked
    #up
    temp = neighbour(x, y - 1, 0, img_size[1], coordinates)
    if temp:
        neighbours.append(temp)

    #right höger
    temp = neighbour(x + 1, y, 1, img_size[0], coordinates)
    if temp:
        neighbours.append(temp)

    #down
    temp = neighbour(x, y + 1, 0, img_size[1], coordinates)
    if temp:
        neighbours.append(temp)

    #left vänster
    temp = neighbour(x - 1, y, 1, img_size[0], coordinates)
    if temp:
        neighbours.append(temp)

    return neighbours

def explore_bounding_box(point, pixel_coordinates_list, img_size):
    bounding_box = BoundingBox(point, point)
    queue = []
    pixel_coordinates_list[find_in_list(point, pixel_coordinates_list)][2] = True
    queue.append(point)

    while len(queue) > 0:
        item = queue.pop(0)

        bounding_box.expand_box(item)

        for neighbour in get_neighbours(item, pixel_coordinates_list, img_size):
            if neighbour[2] != True:
                pixel_coordinates_list[find_in_list(neighbour[0], pixel_coordinates_list)][2] = True
                queue.append(neighbour[0])

    return bounding_box


def main():
    path = 'hus.png'

    image = Image.open(path)
    #image.convert('L')
    pixles = list(image.getdata())
    pixel_coordinates = []
    visted = False
    width = image.width

    print("divmod")
    for i, pixel in enumerate(pixles):
        pixel_coordinates.append([list(divmod(i, width)), (pixel), [visted]])

    print("loop image")
    box = []
    for i, item in enumerate(pixel_coordinates):
        if item[1] != (0,0,0,0) and len([i for i in box if i.in_box(item[0]) == True]) == 0:
            print(item)
            box.append(explore_bounding_box(pixel_coordinates[i][0], pixel_coordinates, image.size))


    print("draw")
    draw = ImageDraw.Draw(image)
    for i in box:    
        print("box", i.getRect())
        draw.rectangle(i.getRect(), fill =None, outline ="red")
        image.save("recttest.png", "png")

if __name__ == '__main__':
    main()
