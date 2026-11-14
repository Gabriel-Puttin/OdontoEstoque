package services

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"odontoestoque/internal/models"
)

type ProductRepository interface {
	FindAll(ctx context.Context) ([]models.Product, error)
	FindByID(ctx context.Context, id int64) (*models.Product, error)
	Create(ctx context.Context, in models.ProductInput, quantidade int32, usuarioID *int64) (int64, error)
	Update(ctx context.Context, id int64, in models.ProductInput, novaQuantidade *int32, usuarioID *int64) error
	Delete(ctx context.Context, id int64) error
}

type ProductService struct {
	repo ProductRepository
}

func NewProductService(repo ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) List(ctx context.Context) ([]models.Product, error) {
	return s.repo.FindAll(ctx)
}

func (s *ProductService) Get(ctx context.Context, id int64) (*models.Product, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *ProductService) Create(ctx context.Context, req models.ProductRequest) (*models.Product, error) {
	in, err := validateProduct(req)
	if err != nil {
		return nil, err
	}
	var quantidade int32
	if req.Quantidade != nil {
		quantidade = *req.Quantidade
	}
	id, err := s.repo.Create(ctx, in, quantidade, req.UsuarioID)
	if err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, id)
}

func (s *ProductService) Update(ctx context.Context, id int64, req models.ProductRequest) (*models.Product, error) {
	in, err := validateProduct(req)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, id, in, req.Quantidade, req.UsuarioID); err != nil {
		return nil, err
	}
	return s.repo.FindByID(ctx, id)
}

func (s *ProductService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

// validateProduct confere formato e limites das colunas. Regras que dependem do
// tipo do produto (validade obrigatória/proibida, variação) ficam com o trigger do banco.
func validateProduct(req models.ProductRequest) (models.ProductInput, error) {
	fields := map[string]string{}
	in := models.ProductInput{
		TipoProdutoCodigo: strings.ToUpper(strings.TrimSpace(req.TipoProdutoCodigo)),
		CategoriaID:       req.CategoriaID,
		UnidadeMedidaID:   req.UnidadeMedidaID,
		Nome:              strings.TrimSpace(req.Nome),
	}

	if in.Nome == "" {
		fields["nome"] = "obrigatório"
	} else if utf8.RuneCountInString(in.Nome) > 120 {
		fields["nome"] = "máximo de 120 caracteres"
	}

	if req.Variacao != nil {
		if v := strings.TrimSpace(*req.Variacao); v != "" {
			if utf8.RuneCountInString(v) > 40 {
				fields["variacao"] = "máximo de 40 caracteres"
			}
			in.Variacao = &v
		}
	}

	if in.TipoProdutoCodigo == "" {
		fields["tipo_produto_codigo"] = "obrigatório"
	}
	if in.CategoriaID <= 0 {
		fields["categoria_id"] = "obrigatório"
	}
	if in.UnidadeMedidaID <= 0 {
		fields["unidade_medida_id"] = "obrigatório"
	}

	if req.EstoqueMinimo == nil {
		fields["estoque_minimo"] = "obrigatório"
	} else if *req.EstoqueMinimo < 0 {
		fields["estoque_minimo"] = "não pode ser negativo"
	} else {
		in.EstoqueMinimo = *req.EstoqueMinimo
	}

	if req.Quantidade != nil && *req.Quantidade < 0 {
		fields["quantidade"] = "não pode ser negativa"
	}

	if req.Validade != nil && strings.TrimSpace(*req.Validade) != "" {
		d, err := time.Parse(time.DateOnly, strings.TrimSpace(*req.Validade))
		if err != nil {
			fields["validade"] = "use o formato AAAA-MM-DD"
		} else {
			in.Validade = &d
		}
	}

	if len(fields) > 0 {
		return in, &models.ValidationError{Fields: fields}
	}
	return in, nil
}
