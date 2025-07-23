from PIL import Image, ImageFont, ImageDraw
import os

font_path = "./output3/"
font_name = "Hack-Regular.ttf"
out_path = font_path
font_size = 14
font_color = "#FFFFFF"
background_color = "#000000"

img_width = 9  
img_height = 14  

no_center_chars = {'"', '.', '_', '^', ','}

os.makedirs(out_path, exist_ok=True)

try:
    font = ImageFont.truetype(font_name, font_size)
except OSError:
    print(f"Font file '{font_name}' not found. Using default font.")
    font = ImageFont.load_default()

desired_characters = "!\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"

print(f"Extracting {len(desired_characters)} characters...")
print(f"All images will be {img_width}x{img_height} pixels")
print("Character | ASCII | Char Width | Char Height | Centered")
print("-" * 55)

for i, character in enumerate(desired_characters):
    try:
        bbox = font.getbbox(character)
        left, top, right, bottom = bbox
        
        char_width = right - left
        char_height = bottom - top
        
        img = Image.new("RGB", (img_width, img_height), background_color)
        draw = ImageDraw.Draw(img)
        
        if character in no_center_chars:
            x_pos = (img_width - char_width) // 2 - left 
            if character == '_':
                y_pos = img_height - 5 - top
            elif character == ',':
                y_pos = img_height - 6 - top
            elif character == '.':
                y_pos = img_height - 5 - top
                y_pos = 1 - top
            elif character == '"':
                y_pos = 1 - top
            else:
                y_pos = -top
            centered_status = "No"
        else:
            x_pos = (img_width - char_width) // 2 - left
            y_pos = (img_height - char_height) // 2 - top
            centered_status = "Yes"
        
        draw.text((x_pos, y_pos), character, font=font, fill=font_color)
        
        ascii_code = ord(character)
        filename = f"{ascii_code}.png"
        filepath = os.path.join(out_path, filename)
        img.save(filepath)
        
        print(f"    {character}     | {ascii_code:3d}  |     {char_width:2d}     |     {char_height:2d}      |   {centered_status}")
        
    except Exception as e:
        print(f"[-] Couldn't save character '{character}': {e}")

print(f"\nExtraction complete! Files saved to: {out_path}")
print(f"Successfully extracted {len(desired_characters)} character images.")
print(f"All images are {img_width}x{img_height} pixels.")
print(f"Characters {no_center_chars} were NOT centered.")