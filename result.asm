; test-asm
bits 16
mov cx, 4
mov ax, 10
add ax, 10
cmp cx, 4
loopz $+2+-8
