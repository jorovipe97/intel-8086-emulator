file=listing_0041_add_sub_cmp_jnz

go run main.go $file

echo "Assemblying resulting file..."
./nasm result.asm
echo "Successful!"

echo "Assemblying original file..."
./nasm listings/$file.asm
echo "Succesfull!"

echo "Comparing both binaries..."
cmp listings/$file result
