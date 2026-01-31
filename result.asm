; listing_0041_add_sub_cmp_jnz
bits 16
add bx, [bx+si]
add bx, [bp]
add si, 2
add bp, 2
add cx, 8
add bx, [bp]
add cx, [bx+2]
add bh, [bp+si+4]
add di, [bp+di+6]
add word [bx+si], bx
add word [bp], bx
add word [bp], bx
add word [bx+2], cx
add byte [bp+si+4], bh
add word [bp+di+6], di
add byte [bx], 34
add word [bp+si+1000], 29
add ax, [bp]
add al, [bx+si]
add ax, bx
add al, ah
add ax, 1000
add al, -30
add al, 9
sub bx, [bx+si]
sub bx, [bp]
sub si, 2
sub bp, 2
sub cx, 8
sub bx, [bp]
sub cx, [bx+2]
sub bh, [bp+si+4]
sub di, [bp+di+6]
sub word [bx+si], bx
sub word [bp], bx
sub word [bp], bx
sub word [bx+2], cx
sub byte [bp+si+4], bh
sub word [bp+di+6], di
sub byte [bx], 34
sub word [bx+di], 29
sub ax, [bp]
sub al, [bx+si]
sub ax, bx
sub al, ah
sub ax, 1000
sub al, -30
sub al, 9
cmp bx, [bx+si]
cmp bx, [bp]
cmp si, 2
cmp bp, 2
cmp cx, 8
cmp bx, [bp]
cmp cx, [bx+2]
cmp bh, [bp+si+4]
cmp di, [bp+di+6]
cmp word [bx+si], bx
cmp word [bp], bx
cmp word [bp], bx
cmp word [bx+2], cx
cmp byte [bp+si+4], bh
cmp word [bp+di+6], di
cmp byte [bx], 34
cmp word [+4834], 29
cmp ax, [bp]
cmp al, [bx+si]
cmp ax, bx
cmp al, ah
cmp ax, 1000
cmp al, -30
cmp al, 9
