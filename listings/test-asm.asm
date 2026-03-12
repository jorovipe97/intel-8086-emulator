bits 16

mov ax, 12
mov bx, 12
mov cx, 20
cmp ax, bx
jnz not_equal
; If not equal cx will end up 25 = 20 + 5
; If equal cx will end up 30 = 20 + 5 + 5
add cx, 5
not_equal:
add cx, 5
