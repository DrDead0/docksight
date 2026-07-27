import { ApiProperty } from '@nestjs/swagger';
import { IsEmail, IsString, MinLength } from 'class-validator';

/**
 * First-run admin creation. Password length is enforced here so a bad request
 * is rejected at the edge with a 400, before any hashing work happens.
 */
export class CreateAdminDto {
  @ApiProperty({ example: 'admin@example.com' })
  @IsEmail({}, { message: 'A valid email is required' })
  email!: string;

  @ApiProperty({ example: 'super-secret-password', minLength: 8 })
  @IsString()
  @MinLength(8, { message: 'Password must be at least 8 characters' })
  password!: string;
}
