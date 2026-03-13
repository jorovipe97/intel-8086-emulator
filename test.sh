file=listing_0050_challenge_jumps

echo "Assemblying original file..."
./nasm listings/$file.asm
echo "Succesfull!"

go run main.go $file --simulate

echo "Assemblying resulting file..."
./nasm result.asm
echo "Successful!"

echo "Comparing both binaries..."
cmp listings/$file result
