  .org $0000

  .org $0200
  ldx #$00
  ldy #$00
  jsr sub1
  jsr sub2
  jsr sub1
  jsr sub2
  jsr sub1
  txa
  tya
  .byte 02

  .org $0300
sub1:
    inx
    rts

  .org $0400
sub2:
    iny
    rts

