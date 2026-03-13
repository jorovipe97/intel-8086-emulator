; test-asm
bits 16
mov ax, 12
mov bx, 12
mov cx, 20
cmp ax, bx
jnz $+2+3
add cx, 5
add cx, 5
