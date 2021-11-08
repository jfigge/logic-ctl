  .org $8000
  lda #$ff
  sta $6002
  lda #$E0
  sta $6003
reset:
  ldx #$08
loop:
  lda data,x
  sta $6000
  dex
  bne loop
  jmp reset

  .org $805c
data:
  .byte $01
  .byte $02
  .byte $04
  .byte $08
  .byte $0f
  .byte $10
  .byte $20
  .byte $40
  .byte $80
  .byte $f0

