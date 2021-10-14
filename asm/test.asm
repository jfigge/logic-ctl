  .org $0
  lda #%11110000  ; set 8-bit mode; 2-line display; 5x8 font
  lda $FF
  lda #$FF
  jsr $55
  word "h"

