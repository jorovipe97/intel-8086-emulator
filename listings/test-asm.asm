bits 16

mov cx, 4
mov ax, 10

my_loop:
add ax, 10
cmp cx, 4
loopz my_loop
