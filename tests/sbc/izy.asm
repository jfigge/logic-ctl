  .org $0000

  .org $0200
  sec
  lda #$ff
  ldy #$01
  sbc ($10),y
  iny
  sbc ($10),y
  iny
  sbc ($10),y
  iny
  sbc ($10),y
  iny
  sbc ($10),y

  .org $0010
  .word $02fc

  .org $02fd
  .byte $03
  .byte $fc
  .byte $00
  .byte $01
  .byte $fe
