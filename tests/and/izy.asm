  .org $0000

  .org $0200
  lda #$ff
  ldy #$01
  and ($10),y
  iny
  and ($10),y
  iny
  and ($10),y
  iny
  and ($10),y
  iny
  and ($10),y

  .org $0010
  .word $02fc

  .org $02fd
  .byte $ff
  .byte $0f
  .byte $ff
  .byte $f8
  .byte $00
