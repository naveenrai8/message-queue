package com.nr.messagequeuesql.dto;

import jakarta.validation.constraints.NotBlank;
import lombok.Builder;

@Builder
public record MessageRequestDto (
        @NotBlank
        String message
) {
}
