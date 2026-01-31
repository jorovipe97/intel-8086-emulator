file=listing_0041_add_sub_cmp_jnz

echo "Assemblying original file..."
./nasm listings/$file.asm
echo "Succesfull!"

go run main.go $file

echo "Assemblying resulting file..."
./nasm result.asm
echo "Successful!"

echo "Comparing both binaries..."
cmp listings/$file result
