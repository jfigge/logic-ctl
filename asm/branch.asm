  ldx  #$02
loop:
  dex
  bne  loop
  jmp  $8000
